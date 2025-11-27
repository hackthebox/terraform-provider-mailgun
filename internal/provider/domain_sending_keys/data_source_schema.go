// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// DomainSendingKeysDataSourceSchema returns the schema for the data source.
func DomainSendingKeysDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Lists all domain sending keys for a specific Mailgun domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain to list sending keys for.",
				Required:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of sending keys for this domain.",
				Computed:    true,
			},
			"keys": schema.ListNestedAttribute{
				Description: "List of domain sending keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the sending key.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Human-readable description of the key.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the key was created.",
							Computed:    true,
						},
						"expires_at": schema.StringAttribute{
							Description: "Timestamp when the key expires, if applicable.",
							Computed:    true,
						},
						"is_disabled": schema.BoolAttribute{
							Description: "Whether the key is disabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
