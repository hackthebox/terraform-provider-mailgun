// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// IPAllowlistResourceSchema returns the schema for the mailgun_ip_allowlist resource.
func IPAllowlistResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Mailgun IP allowlist entry. IP allowlisting restricts API key and SMTP credential usage to specific IP addresses.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "The IP address or CIDR range to allowlist (e.g., '192.168.1.1' or '192.168.1.0/24').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for this IP allowlist entry.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the IP allowlist entry (same as the address).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
