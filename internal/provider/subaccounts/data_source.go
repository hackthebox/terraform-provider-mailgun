// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package subaccounts

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &subaccountDataSource{}
	_ datasource.DataSourceWithConfigure = &subaccountDataSource{}
)

// NewSubaccountDataSource creates a new subaccount data source.
func NewSubaccountDataSource() datasource.DataSource {
	return &subaccountDataSource{}
}

type subaccountDataSource struct {
	client *mailgun.Client
}

func (d *subaccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

func (d *subaccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SubaccountDataSourceSchema()
}

func (d *subaccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *subaccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SubaccountModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccountID := state.ID.ValueString()

	// Get subaccount from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	subaccount, err := d.client.GetSubaccount(readCtx, subaccountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Subaccount",
			fmt.Sprintf("Could not read subaccount %s: %s", subaccountID, err.Error()),
		)
		return
	}

	// Map response to state
	state.ID = types.StringValue(subaccount.Item.ID)
	state.Name = types.StringValue(subaccount.Item.Name)
	state.Status = types.StringValue(subaccount.Item.Status)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
