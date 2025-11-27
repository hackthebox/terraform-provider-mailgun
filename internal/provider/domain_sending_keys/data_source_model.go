// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainSendingKeysDataSourceModel represents the Terraform state for the data source.
type DomainSendingKeysDataSourceModel struct {
	Domain     types.String `tfsdk:"domain"`
	TotalCount types.Int64  `tfsdk:"total_count"`
	Keys       types.List   `tfsdk:"keys"` // List of DomainSendingKeyItemModel
}

// DomainSendingKeyItemModel represents a single key item in the data source.
type DomainSendingKeyItemModel struct {
	Id          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	IsDisabled  types.Bool   `tfsdk:"is_disabled"`
}
