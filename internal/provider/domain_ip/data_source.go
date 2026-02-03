// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &domainIPsDataSource{}
	_ datasource.DataSourceWithConfigure = &domainIPsDataSource{}
)

// NewDomainIPsDataSource creates a new domain IPs data source.
func NewDomainIPsDataSource() datasource.DataSource {
	return &domainIPsDataSource{}
}

type domainIPsDataSource struct {
	client *mailgun.Client
}

func (d *domainIPsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_ips"
}

func (d *domainIPsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainIPsDataSourceSchema()
}

func (d *domainIPsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainIPsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DomainIPsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List all IPs for the domain
	ips, err := d.client.ListDomainIPs(readCtx, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain IPs",
			fmt.Sprintf("Could not read IPs for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Map to state
	state.IPs = make([]IPItemModel, len(ips))
	for i, ip := range ips {
		state.IPs[i] = IPItemModel{
			IP: types.StringValue(ip.IP),
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
