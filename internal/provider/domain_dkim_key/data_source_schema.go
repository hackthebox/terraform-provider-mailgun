// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// DomainDkimKeysDataSourceSchema returns the schema for the mailgun_domain_dkim_keys data source.
func DomainDkimKeysDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves DKIM keys for a Mailgun domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to retrieve DKIM keys for.",
				Required:    true,
			},
			"keys": schema.ListNestedAttribute{
				Description: "List of DKIM keys for the domain.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"selector": schema.StringAttribute{
							Description: "The DKIM selector.",
							Computed:    true,
						},
						"signing_domain": schema.StringAttribute{
							Description: "The domain used for signing.",
							Computed:    true,
						},
						"active": schema.BoolAttribute{
							Description: "Whether the DKIM key is active.",
							Computed:    true,
						},
						"dns_record_name": schema.StringAttribute{
							Description: "The DNS record name.",
							Computed:    true,
						},
						"dns_record_type": schema.StringAttribute{
							Description: "The DNS record type.",
							Computed:    true,
						},
						"dns_record_value": schema.StringAttribute{
							Description: "The DNS record value.",
							Computed:    true,
						},
						"dns_record_valid": schema.StringAttribute{
							Description: "Whether the DNS record is valid.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
