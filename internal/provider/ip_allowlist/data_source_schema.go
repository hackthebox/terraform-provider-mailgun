// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// IPAllowlistDataSourceSchema returns the schema for the mailgun_ip_allowlist data source.
func IPAllowlistDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves all IP allowlist entries for the Mailgun account. IP allowlisting restricts API key and SMTP credential usage to specific IP addresses.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this data source.",
				Computed:    true,
			},
			"entries": schema.ListNestedAttribute{
				Description: "List of IP allowlist entries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "The IP address or CIDR range that is allowlisted.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the IP allowlist entry.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
