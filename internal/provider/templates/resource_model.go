// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TemplateModel represents the Terraform state for a Mailgun template.
type TemplateModel struct {
	// Required fields
	Domain types.String `tfsdk:"domain"`
	Name   types.String `tfsdk:"name"`

	// Optional fields
	Description types.String `tfsdk:"description"`

	// Version fields (for initial version)
	Template types.String `tfsdk:"template"`
	Engine   types.String `tfsdk:"engine"`
	Comment  types.String `tfsdk:"comment"`

	// Computed fields
	Id        types.String `tfsdk:"id"` // Composite ID: domain/name
	CreatedAt types.String `tfsdk:"created_at"`

	// Computed version fields
	VersionTag    types.String `tfsdk:"version_tag"`
	VersionActive types.Bool   `tfsdk:"version_active"`
}
