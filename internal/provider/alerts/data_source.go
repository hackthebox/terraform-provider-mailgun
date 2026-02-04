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
	_ datasource.DataSource              = &alertDataSource{}
	_ datasource.DataSourceWithConfigure = &alertDataSource{}
)

// NewAlertDataSource creates a new alert data source.
func NewAlertDataSource() datasource.DataSource {
	return &alertDataSource{}
}

type alertDataSource struct {
	client *mailgun.Client
}

func (d *alertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (d *alertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AlertDataSourceSchema()
}

func (d *alertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AlertModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertID := state.ID.ValueString()

	// Get alerts from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	alertsResp, err := d.client.ListAlerts(readCtx, &mailgun.ListAlertsOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Alert",
			fmt.Sprintf("Could not read alert %s: %s", alertID, err.Error()),
		)
		return
	}

	// Find the alert in the list
	var found bool
	for _, alert := range alertsResp.Events {
		if alert.ID != nil && alert.ID.String() == alertID {
			found = true

			// Update state from API response
			state.EventType = types.StringValue(alert.EventType)
			state.Channel = types.StringValue(string(alert.Channel))

			// Map channel settings
			switch alert.Channel {
			case "email":
				if len(alert.Settings.Emails) > 0 {
					emails, diags := types.ListValueFrom(ctx, types.StringType, alert.Settings.Emails)
					resp.Diagnostics.Append(diags...)
					state.Emails = emails
				} else {
					state.Emails = types.ListNull(types.StringType)
				}
				state.WebhookURL = types.StringNull()
				state.SlackIDs = types.ListNull(types.StringType)

			case "webhook":
				if alert.Settings.URL != nil {
					state.WebhookURL = types.StringValue(*alert.Settings.URL)
				} else {
					state.WebhookURL = types.StringNull()
				}
				state.Emails = types.ListNull(types.StringType)
				state.SlackIDs = types.ListNull(types.StringType)

			case "slack":
				if len(alert.Settings.ChannelIDs) > 0 {
					slackIDs, diags := types.ListValueFrom(ctx, types.StringType, alert.Settings.ChannelIDs)
					resp.Diagnostics.Append(diags...)
					state.SlackIDs = slackIDs
				} else {
					state.SlackIDs = types.ListNull(types.StringType)
				}
				state.Emails = types.ListNull(types.StringType)
				state.WebhookURL = types.StringNull()

			default:
				state.Emails = types.ListNull(types.StringType)
				state.WebhookURL = types.StringNull()
				state.SlackIDs = types.ListNull(types.StringType)
			}

			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Alert Not Found",
			fmt.Sprintf("Alert with ID '%s' was not found.", alertID),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
