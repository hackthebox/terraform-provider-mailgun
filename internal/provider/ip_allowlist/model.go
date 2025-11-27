// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IPAllowlistModel represents the Terraform state for a Mailgun IP allowlist entry.
type IPAllowlistModel struct {
	// Required fields
	Address types.String `tfsdk:"address"`

	// Optional fields
	Description types.String `tfsdk:"description"`

	// Computed fields
	Id types.String `tfsdk:"id"` // Same as address (IP address is the unique identifier)
}

// IPAllowlistDataSourceModel represents a single entry in the data source.
type IPAllowlistEntryModel struct {
	Address     types.String `tfsdk:"address"`
	Description types.String `tfsdk:"description"`
}

// IPAllowlistListDataSourceModel represents the data source for listing all IP allowlist entries.
type IPAllowlistListDataSourceModel struct {
	// Computed fields
	Id      types.String            `tfsdk:"id"`
	Entries []IPAllowlistEntryModel `tfsdk:"entries"`
}
