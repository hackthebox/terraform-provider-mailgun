// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ datasource.DataSource              = &domainDkimKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &domainDkimKeysDataSource{}
)

// NewDomainDkimKeysDataSource creates a new domain DKIM keys data source.
func NewDomainDkimKeysDataSource() datasource.DataSource {
	return &domainDkimKeysDataSource{}
}

type domainDkimKeysDataSource struct {
	client *mailgun.Client
}

func (d *domainDkimKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dkim_keys"
}

func (d *domainDkimKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainDkimKeysDataSourceSchema()
}

func (d *domainDkimKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *domainDkimKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DomainDkimKeysDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List all DKIM keys for the domain
	iter := d.client.ListDomainKeys(domain)

	var allKeys []mtypes.DomainKey
	var keys []mtypes.DomainKey
	for iter.Next(readCtx, &keys) {
		allKeys = append(allKeys, keys...)
	}

	if err := iter.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain DKIM Keys",
			fmt.Sprintf("Could not read DKIM keys for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Map to state
	state.Keys = make([]DkimKeyItemModel, len(allKeys))
	for i, key := range allKeys {
		state.Keys[i] = DkimKeyItemModel{
			Selector:       types.StringValue(key.Selector),
			SigningDomain:  types.StringValue(key.SigningDomain),
			Active:         types.BoolValue(key.DNSRecord.Active),
			DnsRecordName:  types.StringValue(key.DNSRecord.Name),
			DnsRecordType:  types.StringValue(key.DNSRecord.RecordType),
			DnsRecordValue: types.StringValue(key.DNSRecord.Value),
			DnsRecordValid: types.StringValue(key.DNSRecord.Valid),
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
