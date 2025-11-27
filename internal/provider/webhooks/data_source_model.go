// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WebhooksDataSourceModel represents the Terraform state for the mailgun_webhooks data source.
type WebhooksDataSourceModel struct {
	Domain     types.String `tfsdk:"domain"`
	TotalCount types.Int64  `tfsdk:"total_count"`
	Webhooks   types.List   `tfsdk:"webhooks"` // List of WebhookItemModel
}

// WebhookItemModel represents a single webhook item in the data source.
type WebhookItemModel struct {
	WebhookType types.String `tfsdk:"webhook_type"`
	Urls        types.List   `tfsdk:"urls"` // List of strings
}
