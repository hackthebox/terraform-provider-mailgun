// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TemplateVersionModel represents the Terraform state for a Mailgun template version.
type TemplateVersionModel struct {
	// Required fields
	Domain       types.String `tfsdk:"domain"`
	TemplateName types.String `tfsdk:"template_name"`
	Template     types.String `tfsdk:"template"`

	// Optional fields
	Tag     types.String `tfsdk:"tag"`
	Engine  types.String `tfsdk:"engine"`
	Comment types.String `tfsdk:"comment"`
	Active  types.Bool   `tfsdk:"active"`

	// Computed fields
	Id        types.String `tfsdk:"id"` // Composite ID: domain/template_name/tag
	CreatedAt types.String `tfsdk:"created_at"`
}
