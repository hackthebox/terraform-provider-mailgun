// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &ipAllowlistDataSource{}
	_ datasource.DataSourceWithConfigure = &ipAllowlistDataSource{}
)

// NewIPAllowlistDataSource creates a new IP allowlist data source.
func NewIPAllowlistDataSource() datasource.DataSource {
	return &ipAllowlistDataSource{}
}

type ipAllowlistDataSource struct {
	client *IPAllowlistClient
}

func (d *ipAllowlistDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_allowlist"
}

func (d *ipAllowlistDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = IPAllowlistDataSourceSchema()
}

func (d *ipAllowlistDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	mg, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = NewIPAllowlistClient(mg)
}

func (d *ipAllowlistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IPAllowlistListDataSourceModel

	// Get all IP allowlist entries from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	entries, err := d.client.ListIPAllowlist(readCtx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IP Allowlist",
			fmt.Sprintf("Could not read IP allowlist entries: %s", err.Error()),
		)
		return
	}

	// Map API response to state
	state.Entries = make([]IPAllowlistEntryModel, len(entries))
	for i, entry := range entries {
		state.Entries[i] = IPAllowlistEntryModel{
			Address:     types.StringValue(entry.IPAddress),
			Description: types.StringValue(entry.Description),
		}
	}

	// Set a stable ID for the data source
	state.Id = types.StringValue("ip_allowlist")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
