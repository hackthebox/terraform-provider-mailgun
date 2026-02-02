// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TemplatesDataSourceModel represents the data source state for listing templates.
type TemplatesDataSourceModel struct {
	Domain     types.String        `tfsdk:"domain"`
	TotalCount types.Int64         `tfsdk:"total_count"`
	Templates  []TemplateItemModel `tfsdk:"templates"`
}

// TemplateItemModel represents a single template in the list.
type TemplateItemModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	CreatedAt     types.String `tfsdk:"created_at"`
	VersionTag    types.String `tfsdk:"version_tag"`
	VersionEngine types.String `tfsdk:"version_engine"`
	VersionActive types.Bool   `tfsdk:"version_active"`
}
