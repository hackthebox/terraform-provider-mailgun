// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ resource.Resource              = &alertResource{}
	_ resource.ResourceWithConfigure = &alertResource{}
)

// NewAlertResource creates a new alert resource.
func NewAlertResource() resource.Resource {
	return &alertResource{}
}

type alertResource struct {
	client *mailgun.Client
}

func (r *alertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *alertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AlertResourceSchema()
}

func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build channel settings based on channel type
	settings := mtypes.AlertsChannelSettings{}
	channel := plan.Channel.ValueString()

	switch channel {
	case mtypes.AlertsEmailChannel:
		if plan.Emails.IsNull() || len(plan.Emails.Elements()) == 0 {
			resp.Diagnostics.AddError(
				"Missing Required Attribute",
				"The 'emails' attribute is required when channel is 'email'.",
			)
			return
		}
		var emails []string
		diags = plan.Emails.ElementsAs(ctx, &emails, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		settings.Emails = emails

	case mtypes.AlertsWebhookChannel:
		if plan.WebhookURL.IsNull() || plan.WebhookURL.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing Required Attribute",
				"The 'webhook_url' attribute is required when channel is 'webhook'.",
			)
			return
		}
		url := plan.WebhookURL.ValueString()
		settings.URL = &url

	case mtypes.AlertsSlackChannel:
		if plan.SlackIDs.IsNull() || len(plan.SlackIDs.Elements()) == 0 {
			resp.Diagnostics.AddError(
				"Missing Required Attribute",
				"The 'slack_ids' attribute is required when channel is 'slack'.",
			)
			return
		}
		var slackIDs []string
		diags = plan.SlackIDs.ElementsAs(ctx, &slackIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		settings.ChannelIDs = slackIDs
	}

	// Create the alert request
	// Note: We use a switch to assign the channel because the SDK uses a named string type
	// from an internal package that we cannot import directly
	var alertReq mtypes.AlertsEventSettingRequest
	alertReq.EventType = plan.EventType.ValueString()
	alertReq.Settings = settings

	switch channel {
	case mtypes.AlertsEmailChannel:
		alertReq.Channel = mtypes.AlertsEmailChannel
	case mtypes.AlertsWebhookChannel:
		alertReq.Channel = mtypes.AlertsWebhookChannel
	case mtypes.AlertsSlackChannel:
		alertReq.Channel = mtypes.AlertsSlackChannel
	}

	// Make API call
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	alertResp, err := r.client.AddAlert(createCtx, alertReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Alert",
			fmt.Sprintf("Could not create alert: %s", err.Error()),
		)
		return
	}

	// Map response to state
	if alertResp.ID != nil {
		plan.ID = types.StringValue(alertResp.ID.String())
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the alert ID
	alertID := state.ID.ValueString()

	// List all alerts and find the one we're looking for
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	alertsResp, err := r.client.ListAlerts(readCtx, &mailgun.ListAlertsOptions{})
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
				}
				state.WebhookURL = types.StringNull()
				state.SlackIDs = types.ListNull(types.StringType)

			case "webhook":
				if alert.Settings.URL != nil {
					state.WebhookURL = types.StringValue(*alert.Settings.URL)
				}
				state.Emails = types.ListNull(types.StringType)
				state.SlackIDs = types.ListNull(types.StringType)

			case "slack":
				if len(alert.Settings.ChannelIDs) > 0 {
					slackIDs, diags := types.ListValueFrom(ctx, types.StringType, alert.Settings.ChannelIDs)
					resp.Diagnostics.Append(diags...)
					state.SlackIDs = slackIDs
				}
				state.Emails = types.ListNull(types.StringType)
				state.WebhookURL = types.StringNull()
			}

			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Alerts are immutable - all attributes have RequiresReplace
	// This method should never be called, but we implement it to satisfy the interface
	resp.Diagnostics.AddError(
		"Alert Update Not Supported",
		"Mailgun alerts cannot be updated. Any change requires destroying and recreating the resource.",
	)
}

func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertID := state.ID.ValueString()

	// Parse UUID
	alertUUID, err := uuid.Parse(alertID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Alert ID",
			fmt.Sprintf("Could not parse alert ID '%s' as UUID: %s", alertID, err.Error()),
		)
		return
	}

	// Delete the alert
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = r.client.DeleteAlert(deleteCtx, alertUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Mailgun Alert",
			fmt.Sprintf("Could not delete alert %s: %s", alertID, err.Error()),
		)
		return
	}
}
