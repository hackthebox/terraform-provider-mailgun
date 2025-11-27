// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WebhookModel represents the Terraform state for a Mailgun webhook.
type WebhookModel struct {
	// Required fields
	Domain      types.String `tfsdk:"domain"`
	WebhookType types.String `tfsdk:"webhook_type"`
	Urls        types.List   `tfsdk:"urls"` // List of strings (max 3)

	// Computed fields
	Id types.String `tfsdk:"id"` // Composite ID: domain/webhook_type
}
