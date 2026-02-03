// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// MailingListMembersDataSourceSchema returns the schema for the mailgun_mailing_list_members data source.
func MailingListMembersDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves a list of all members in a Mailgun mailing list.",
		Attributes: map[string]schema.Attribute{
			"list_address": schema.StringAttribute{
				Description: "The email address of the mailing list.",
				Required:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "The total number of members found.",
				Computed:    true,
			},
			"members": schema.ListNestedAttribute{
				Description: "List of members.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "The email address of the member.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the member.",
							Computed:    true,
						},
						"subscribed": schema.BoolAttribute{
							Description: "Whether the member is subscribed.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
