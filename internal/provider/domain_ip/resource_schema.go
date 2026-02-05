// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// DomainIPResourceSchema returns the schema for the mailgun_domain_ip resource.
func DomainIPResourceSchema() schema.Schema {
	return schema.Schema{
		Version:     0,
		Description: "Manages IP address associations for a Mailgun domain. This resource allows you to assign dedicated IP addresses to your domains for sending emails.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this domain IP association (domain/ip).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The domain name to associate the IP with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Description: "The IP address to associate with the domain.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
