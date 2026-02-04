// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package subaccounts

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SubaccountModel represents a single subaccount for the data source.
type SubaccountModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

// SubaccountsListModel represents the list data source model.
type SubaccountsListModel struct {
	Enabled     types.Bool        `tfsdk:"enabled"`
	Subaccounts []SubaccountModel `tfsdk:"subaccounts"`
	TotalCount  types.Int64       `tfsdk:"total_count"`
}
