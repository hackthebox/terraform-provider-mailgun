// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainSendingKeyModel represents the Terraform state for a domain sending key.
type DomainSendingKeyModel struct {
	// Required fields
	Domain types.String `tfsdk:"domain"`

	// Optional fields
	Description types.String `tfsdk:"description"`
	Expiration  types.Int64  `tfsdk:"expiration"`

	// Computed fields
	Id        types.String `tfsdk:"id"`
	Secret    types.String `tfsdk:"secret"`
	CreatedAt types.String `tfsdk:"created_at"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}
