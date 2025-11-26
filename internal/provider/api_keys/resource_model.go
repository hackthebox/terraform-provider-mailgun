// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApiKeyModel represents the Terraform state for an API key resource
type ApiKeyModel struct {
	// Required inputs
	Role types.String `tfsdk:"role"`

	// Optional inputs
	Description types.String `tfsdk:"description"`
	DomainName  types.String `tfsdk:"domain_name"`
	Kind        types.String `tfsdk:"kind"`
	Expiration  types.Int64  `tfsdk:"expiration"`

	// Computed attributes
	Id             types.String `tfsdk:"id"`
	Secret         types.String `tfsdk:"secret"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	DisabledReason types.String `tfsdk:"disabled_reason"`
	Requestor      types.String `tfsdk:"requestor"`
	UserName       types.String `tfsdk:"user_name"`
}
