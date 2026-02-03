// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// MailingListsDataSourceSchema returns the schema for the mailgun_mailing_lists data source.
func MailingListsDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves a list of all mailing lists in your Mailgun account.",
		Attributes: map[string]schema.Attribute{
			"total_count": schema.Int64Attribute{
				Description: "The total number of mailing lists found.",
				Computed:    true,
			},
			"mailing_lists": schema.ListNestedAttribute{
				Description: "List of mailing lists.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "The email address of the mailing list.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the mailing list.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the mailing list.",
							Computed:    true,
						},
						"access_level": schema.StringAttribute{
							Description: "The access level of the mailing list.",
							Computed:    true,
						},
						"reply_preference": schema.StringAttribute{
							Description: "The reply preference of the mailing list.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The timestamp when the mailing list was created.",
							Computed:    true,
						},
						"members_count": schema.Int64Attribute{
							Description: "The number of members in the mailing list.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
