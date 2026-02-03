// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TemplateVersionsDataSourceModel represents the data source state for listing template versions.
type TemplateVersionsDataSourceModel struct {
	Domain       types.String               `tfsdk:"domain"`
	TemplateName types.String               `tfsdk:"template_name"`
	TotalCount   types.Int64                `tfsdk:"total_count"`
	Versions     []TemplateVersionItemModel `tfsdk:"versions"`
}

// TemplateVersionItemModel represents a single template version in the list.
type TemplateVersionItemModel struct {
	Tag       types.String `tfsdk:"tag"`
	Engine    types.String `tfsdk:"engine"`
	Comment   types.String `tfsdk:"comment"`
	Active    types.Bool   `tfsdk:"active"`
	CreatedAt types.String `tfsdk:"created_at"`
}
