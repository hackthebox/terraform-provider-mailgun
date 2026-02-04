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
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ datasource.DataSource              = &subaccountsListDataSource{}
	_ datasource.DataSourceWithConfigure = &subaccountsListDataSource{}
)

// NewSubaccountsListDataSource creates a new subaccounts list data source.
func NewSubaccountsListDataSource() datasource.DataSource {
	return &subaccountsListDataSource{}
}

type subaccountsListDataSource struct {
	client *mailgun.Client
}

func (d *subaccountsListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccounts"
}

func (d *subaccountsListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SubaccountsListDataSourceSchema()
}

func (d *subaccountsListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *subaccountsListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SubaccountsListModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build list options
	opts := &mailgun.ListSubaccountsOptions{
		Limit: 100, // Fetch up to 100 at a time
	}

	// Apply enabled filter if specified
	if !state.Enabled.IsNull() {
		opts.Enabled = state.Enabled.ValueBool()
	}

	// Get subaccounts from API
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	iter := d.client.ListSubaccounts(opts)

	var allSubaccounts []mtypes.Subaccount
	var items []mtypes.Subaccount

	// Use First() to get the first page
	if iter.First(readCtx, &items) {
		allSubaccounts = append(allSubaccounts, items...)

		// Continue fetching if there are more pages
		for iter.Next(readCtx, &items) {
			allSubaccounts = append(allSubaccounts, items...)
		}
	}

	if err := iter.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Subaccounts",
			fmt.Sprintf("Could not list subaccounts: %s", err.Error()),
		)
		return
	}

	// Map response to state
	subaccounts := make([]SubaccountModel, 0, len(allSubaccounts))
	for _, sa := range allSubaccounts {
		subaccounts = append(subaccounts, SubaccountModel{
			ID:     types.StringValue(sa.ID),
			Name:   types.StringValue(sa.Name),
			Status: types.StringValue(sa.Status),
		})
	}

	state.Subaccounts = subaccounts
	state.TotalCount = types.Int64Value(int64(len(subaccounts)))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
