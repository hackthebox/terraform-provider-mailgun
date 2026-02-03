// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &domainTrackingDataSource{}
	_ datasource.DataSourceWithConfigure = &domainTrackingDataSource{}
)

// NewDomainTrackingDataSource creates a new domain tracking data source.
func NewDomainTrackingDataSource() datasource.DataSource {
	return &domainTrackingDataSource{}
}

type domainTrackingDataSource struct {
	client *mailgun.Client
}

func (d *domainTrackingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_tracking"
}

func (d *domainTrackingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainTrackingDataSourceSchema()
}

func (d *domainTrackingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainTrackingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DomainTrackingDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tracking, err := d.client.GetDomainTracking(readCtx, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Tracking",
			fmt.Sprintf("Could not read tracking settings for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Map to state
	state.ClickActive = types.BoolValue(tracking.Click.Active)
	state.OpenActive = types.BoolValue(tracking.Open.Active)
	state.UnsubscribeActive = types.BoolValue(tracking.Unsubscribe.Active)

	if tracking.Unsubscribe.HTMLFooter != "" {
		state.UnsubscribeHtmlFooter = types.StringValue(tracking.Unsubscribe.HTMLFooter)
	} else {
		state.UnsubscribeHtmlFooter = types.StringNull()
	}

	if tracking.Unsubscribe.TextFooter != "" {
		state.UnsubscribeTextFooter = types.StringValue(tracking.Unsubscribe.TextFooter)
	} else {
		state.UnsubscribeTextFooter = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
