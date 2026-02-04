// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package alerts

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlertModel represents a Mailgun alert in Terraform state.
type AlertModel struct {
	ID        types.String `tfsdk:"id"`
	EventType types.String `tfsdk:"event_type"`
	Channel   types.String `tfsdk:"channel"`

	// Channel-specific settings (only one should be set based on channel type)
	Emails     types.List   `tfsdk:"emails"`      // For email channel
	WebhookURL types.String `tfsdk:"webhook_url"` // For webhook channel
	SlackIDs   types.List   `tfsdk:"slack_ids"`   // For slack channel
}

// AlertsListModel represents the state for the alerts list data source.
type AlertsListModel struct {
	Alerts     []AlertModel `tfsdk:"alerts"`
	TotalCount types.Int64  `tfsdk:"total_count"`
}
