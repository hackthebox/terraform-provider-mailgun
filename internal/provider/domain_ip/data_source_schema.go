// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// DomainIPsDataSourceSchema returns the schema for the mailgun_domain_ips data source.
func DomainIPsDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves IP addresses associated with a Mailgun domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to retrieve IP addresses for.",
				Required:    true,
			},
			"ips": schema.ListNestedAttribute{
				Description: "List of IP addresses associated with the domain.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Description: "The IP address.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
