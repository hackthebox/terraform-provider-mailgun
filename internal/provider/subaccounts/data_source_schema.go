// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package subaccounts

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// SubaccountDataSourceSchema returns the schema for the mailgun_subaccount data source.
func SubaccountDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Use this data source to look up an existing Mailgun subaccount by ID. " +
			"Subaccounts are child accounts that share the same plan as the primary account but have their own domains, users, and settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the subaccount.",
				Required:    true,
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
	}
}
