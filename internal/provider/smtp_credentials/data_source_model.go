// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SmtpCredentialDataSourceModel represents the Terraform state for a single SMTP credential data source
type SmtpCredentialDataSourceModel struct {
	// Input - required to lookup the credential
	Domain types.String `tfsdk:"domain"`
	Login  types.String `tfsdk:"login"`

	// Computed attributes
	Id        types.String `tfsdk:"id"`
	FullLogin types.String `tfsdk:"full_login"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// SmtpCredentialsListDataSourceModel represents the Terraform state for the SMTP credentials list data source
type SmtpCredentialsListDataSourceModel struct {
	// Input - required to lookup credentials
	Domain types.String `tfsdk:"domain"`

	// Optional filters
	Limit types.Int64 `tfsdk:"limit"`

	// Computed attributes
	Credentials []SmtpCredentialItemModel `tfsdk:"credentials"`
	TotalCount  types.Int64               `tfsdk:"total_count"`
}

// SmtpCredentialItemModel represents a single credential in the list
type SmtpCredentialItemModel struct {
	Login     types.String `tfsdk:"login"`
	FullLogin types.String `tfsdk:"full_login"`
	CreatedAt types.String `tfsdk:"created_at"`
}
