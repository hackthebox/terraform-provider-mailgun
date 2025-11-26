// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoutesDataSourceModel represents the Terraform state for the mailgun_routes data source.
type RoutesDataSourceModel struct {
	Limit      types.Int64 `tfsdk:"limit"`
	TotalCount types.Int64 `tfsdk:"total_count"`
	Routes     types.List  `tfsdk:"routes"` // List of RouteItemModel
}

// RouteItemModel represents a single route item in the data source.
type RouteItemModel struct {
	Id          types.String `tfsdk:"id"`
	Expression  types.String `tfsdk:"expression"`
	Actions     types.List   `tfsdk:"actions"` // List of strings
	Priority    types.Int64  `tfsdk:"priority"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}
