// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package alerts

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlertResourceSchema returns the schema for the mailgun_alert resource.
func AlertResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Mailgun alert. Alerts notify you when specific events occur in your Mailgun account.\n\n" +
			"**⚠️ Important**: Alerts are immutable. Any changes to an alert's configuration will require " +
			"destroying and recreating the resource. This is a limitation of the Mailgun API which does not " +
			"support updating alerts.",
		MarkdownDescription: "Manages a Mailgun alert. Alerts notify you when specific events occur in your Mailgun account.\n\n" +
			"**⚠️ Important**: Alerts are immutable. Any changes to an alert's configuration will require " +
			"destroying and recreating the resource. This is a limitation of the Mailgun API which does not " +
			"support updating alerts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the alert (UUID format).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event_type": schema.StringAttribute{
				Description: "The type of event that triggers the alert. Use the mailgun_alert_events data source " +
					"to get a list of valid event types. Examples: 'hard_bounces', 'complaints', 'unsubscribes'.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"channel": schema.StringAttribute{
				Description: "The delivery channel for the alert. Valid values: 'email', 'webhook', 'slack'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("email", "webhook", "slack"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"emails": schema.ListAttribute{
				Description: "List of email addresses to receive alert notifications. Required when channel is 'email'.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listRequiresReplace{},
				},
			},
			"webhook_url": schema.StringAttribute{
				Description: "The URL to receive webhook notifications. Required when channel is 'webhook'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slack_ids": schema.ListAttribute{
				Description: "List of Slack channel IDs to receive alert notifications. Required when channel is 'slack'. " +
					"Requires Slack integration to be configured in your Mailgun account.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listRequiresReplace{},
				},
			},
		},
	}
}

// listRequiresReplace is a plan modifier that requires replacement when the list changes.
type listRequiresReplace struct{}

func (m listRequiresReplace) Description(_ context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m listRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m listRequiresReplace) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() {
		return
	}

	if req.PlanValue.IsUnknown() {
		return
	}

	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}
