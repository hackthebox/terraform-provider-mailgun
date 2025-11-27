// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &domainSendingKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &domainSendingKeysDataSource{}
)

// NewDomainSendingKeysDataSource creates a new data source for listing domain sending keys.
func NewDomainSendingKeysDataSource() datasource.DataSource {
	return &domainSendingKeysDataSource{}
}

type domainSendingKeysDataSource struct {
	client *mailgun.Client
}

func (d *domainSendingKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_sending_keys"
}

func (d *domainSendingKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainSendingKeysDataSourceSchema()
}

func (d *domainSendingKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainSendingKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainSendingKeysDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := config.Domain.ValueString()

	// List domain sending keys
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	opts := &mailgun.ListAPIKeysOptions{
		DomainName: domain,
		Kind:       "domain",
	}

	keys, err := d.client.ListAPIKeys(readCtx, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Domain Sending Keys",
			fmt.Sprintf("Could not list sending keys for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Filter to only sending keys and map to model
	keyItems := make([]attr.Value, 0)
	for _, key := range keys {
		// Only include sending keys
		if key.Role != "sending" {
			continue
		}

		var expiresAt attr.Value
		if !key.ExpiresAt.IsZero() {
			expiresAt = types.StringValue(key.ExpiresAt.Format(time.RFC3339))
		} else {
			expiresAt = types.StringNull()
		}

		keyObj, objDiags := types.ObjectValue(
			map[string]attr.Type{
				"id":          types.StringType,
				"description": types.StringType,
				"created_at":  types.StringType,
				"expires_at":  types.StringType,
				"is_disabled": types.BoolType,
			},
			map[string]attr.Value{
				"id":          types.StringValue(key.ID),
				"description": types.StringValue(key.Description),
				"created_at":  types.StringValue(key.CreatedAt.Format(time.RFC3339)),
				"expires_at":  expiresAt,
				"is_disabled": types.BoolValue(key.IsDisabled),
			},
		)
		resp.Diagnostics.Append(objDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		keyItems = append(keyItems, keyObj)
	}

	keysList, listDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":          types.StringType,
				"description": types.StringType,
				"created_at":  types.StringType,
				"expires_at":  types.StringType,
				"is_disabled": types.BoolType,
			},
		},
		keyItems,
	)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Keys = keysList
	config.TotalCount = types.Int64Value(int64(len(keyItems)))

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
