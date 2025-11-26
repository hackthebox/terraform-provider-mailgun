// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApiKeyDataSourceModel represents the Terraform state for a single API key data source
type ApiKeyDataSourceModel struct {
	// Input - required to lookup the key
	Id types.String `tfsdk:"id"`

	// Computed attributes
	Role           types.String `tfsdk:"role"`
	Kind           types.String `tfsdk:"kind"`
	Description    types.String `tfsdk:"description"`
	DomainName     types.String `tfsdk:"domain_name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	DisabledReason types.String `tfsdk:"disabled_reason"`
	Requestor      types.String `tfsdk:"requestor"`
	UserName       types.String `tfsdk:"user_name"`
}

// ApiKeysListDataSourceModel represents the Terraform state for the API keys list data source
type ApiKeysListDataSourceModel struct {
	// Optional filters
	DomainName types.String `tfsdk:"domain_name"`
	Kind       types.String `tfsdk:"kind"`

	// Computed attributes
	Keys       []ApiKeyItemModel `tfsdk:"keys"`
	TotalCount types.Int64       `tfsdk:"total_count"`
}

// ApiKeyItemModel represents a single API key in the list
type ApiKeyItemModel struct {
	Id             types.String `tfsdk:"id"`
	Role           types.String `tfsdk:"role"`
	Kind           types.String `tfsdk:"kind"`
	Description    types.String `tfsdk:"description"`
	DomainName     types.String `tfsdk:"domain_name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	DisabledReason types.String `tfsdk:"disabled_reason"`
	Requestor      types.String `tfsdk:"requestor"`
	UserName       types.String `tfsdk:"user_name"`
}
