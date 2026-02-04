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
	_ datasource.DataSource              = &alertsListDataSource{}
	_ datasource.DataSourceWithConfigure = &alertsListDataSource{}
)

// NewAlertsListDataSource creates a new alerts list data source.
func NewAlertsListDataSource() datasource.DataSource {
	return &alertsListDataSource{}
}

type alertsListDataSource struct {
	client *mailgun.Client
}

func (d *alertsListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerts"
}

func (d *alertsListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AlertsListDataSourceSchema()
}

func (d *alertsListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertsListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AlertsListModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get alerts from API
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	alertsResp, err := d.client.ListAlerts(readCtx, &mailgun.ListAlertsOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Alerts",
			fmt.Sprintf("Could not list alerts: %s", err.Error()),
		)
		return
	}

	// Map response to state
	alerts := make([]AlertModel, 0, len(alertsResp.Events))
	for _, alert := range alertsResp.Events {
		alertModel := AlertModel{
			EventType: types.StringValue(alert.EventType),
			Channel:   types.StringValue(string(alert.Channel)),
		}

		if alert.ID != nil {
			alertModel.ID = types.StringValue(alert.ID.String())
		}

		// Map channel settings
		switch alert.Channel {
		case "email":
			if len(alert.Settings.Emails) > 0 {
				emails, diags := types.ListValueFrom(ctx, types.StringType, alert.Settings.Emails)
				resp.Diagnostics.Append(diags...)
				alertModel.Emails = emails
			} else {
				alertModel.Emails = types.ListNull(types.StringType)
			}
			alertModel.WebhookURL = types.StringNull()
			alertModel.SlackIDs = types.ListNull(types.StringType)

		case "webhook":
			if alert.Settings.URL != nil {
				alertModel.WebhookURL = types.StringValue(*alert.Settings.URL)
			} else {
				alertModel.WebhookURL = types.StringNull()
			}
			alertModel.Emails = types.ListNull(types.StringType)
			alertModel.SlackIDs = types.ListNull(types.StringType)

		case "slack":
			if len(alert.Settings.ChannelIDs) > 0 {
				slackIDs, diags := types.ListValueFrom(ctx, types.StringType, alert.Settings.ChannelIDs)
				resp.Diagnostics.Append(diags...)
				alertModel.SlackIDs = slackIDs
			} else {
				alertModel.SlackIDs = types.ListNull(types.StringType)
			}
			alertModel.Emails = types.ListNull(types.StringType)
			alertModel.WebhookURL = types.StringNull()

		default:
			alertModel.Emails = types.ListNull(types.StringType)
			alertModel.WebhookURL = types.StringNull()
			alertModel.SlackIDs = types.ListNull(types.StringType)
		}

		alerts = append(alerts, alertModel)
	}

	state.Alerts = alerts
	state.TotalCount = types.Int64Value(int64(len(alerts)))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
