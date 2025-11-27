// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WebhookResourceSchema returns the schema for the mailgun_webhook resource.
func WebhookResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Mailgun webhook. Webhooks allow you to receive HTTP POST requests for email events like deliveries, opens, clicks, bounces, and complaints.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to configure the webhook for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"webhook_type": schema.StringAttribute{
				Description: "The type of webhook event. Valid values: 'accepted', 'delivered', 'permanent_fail', 'temporary_fail', 'opened', 'clicked', 'unsubscribed', 'complained'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"accepted",
						"delivered",
						"permanent_fail",
						"temporary_fail",
						"opened",
						"clicked",
						"unsubscribed",
						"complained",
					),
				},
			},
			"urls": schema.ListAttribute{
				Description: "List of URLs to receive webhook POST requests. Maximum of 3 URLs allowed.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 3),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the webhook (domain/webhook_type).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
