// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package subaccounts

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// SubaccountsListDataSourceSchema returns the schema for the mailgun_subaccounts data source.
func SubaccountsListDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Use this data source to list all Mailgun subaccounts. " +
			"Subaccounts are child accounts that share the same plan as the primary account but have their own domains, users, and settings.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Filter to only return enabled subaccounts. If not set, returns all subaccounts.",
				Optional:    true,
			},
			"subaccounts": schema.ListNestedAttribute{
				Description: "List of subaccounts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the subaccount.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the subaccount.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The status of the subaccount: 'enabled' or 'disabled'.",
							Computed:    true,
						},
					},
				},
			},
			"total_count": schema.Int64Attribute{
				Description: "The total number of subaccounts returned.",
				Computed:    true,
			},
		},
	}
}
