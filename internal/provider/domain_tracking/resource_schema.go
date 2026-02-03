// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// DomainTrackingResourceSchema returns the schema for the mailgun_domain_tracking resource.
func DomainTrackingResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages tracking settings for a Mailgun domain. This includes click tracking, open tracking, and unsubscribe tracking with customizable footers.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to configure tracking for. Changing this forces recreation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"click_active": schema.BoolAttribute{
				Description: "Whether click tracking is enabled. When enabled, Mailgun rewrites links to track clicks.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"open_active": schema.BoolAttribute{
				Description: "Whether open tracking is enabled. When enabled, Mailgun inserts a tracking pixel to track opens.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"unsubscribe_active": schema.BoolAttribute{
				Description: "Whether unsubscribe tracking is enabled. When enabled, Mailgun adds unsubscribe links to emails.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"unsubscribe_html_footer": schema.StringAttribute{
				Description: "Custom HTML footer for unsubscribe links. Use %%unsubscribe_url%% as placeholder for the unsubscribe link.",
				Optional:    true,
				Computed:    true,
			},
			"unsubscribe_text_footer": schema.StringAttribute{
				Description: "Custom text footer for unsubscribe links. Use %%unsubscribe_url%% as placeholder for the unsubscribe link.",
				Optional:    true,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the tracking configuration (same as domain).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
