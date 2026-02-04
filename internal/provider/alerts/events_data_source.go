// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &alertEventsDataSource{}
	_ datasource.DataSourceWithConfigure = &alertEventsDataSource{}
)

// NewAlertEventsDataSource creates a new alert events data source.
func NewAlertEventsDataSource() datasource.DataSource {
	return &alertEventsDataSource{}
}

type alertEventsDataSource struct {
	client *mailgun.Client
}

// AlertEventsModel represents the state for the alert events data source.
type AlertEventsModel struct {
	Events types.List `tfsdk:"events"`
}

func (d *alertEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_events"
}

func (d *alertEventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AlertEventsDataSourceSchema()
}

func (d *alertEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AlertEventsModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get alert events from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	eventsResp, err := d.client.ListAlertsEvents(readCtx, &mailgun.ListAlertsEventsOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Alert Events",
			fmt.Sprintf("Could not list alert events: %s", err.Error()),
		)
		return
	}

	// Map response to state
	events, diags := types.ListValueFrom(ctx, types.StringType, eventsResp.Events)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Events = events

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
