// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SmtpCredentialModel represents the Terraform state for an SMTP credential resource
type SmtpCredentialModel struct {
	// Required inputs
	Domain types.String `tfsdk:"domain"`
	Login  types.String `tfsdk:"login"`

	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`

	// Computed attributes
	Id        types.String `tfsdk:"id"`
	FullLogin types.String `tfsdk:"full_login"`
	CreatedAt types.String `tfsdk:"created_at"`
}
