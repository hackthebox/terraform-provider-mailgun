// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RouteModel represents the Terraform state for a Mailgun route.
type RouteModel struct {
	// Required fields
	Expression types.String `tfsdk:"expression"`
	Actions    types.List   `tfsdk:"actions"` // List of strings

	// Optional fields
	Priority    types.Int64  `tfsdk:"priority"`
	Description types.String `tfsdk:"description"`

	// Computed fields (read-only)
	Id        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
}
