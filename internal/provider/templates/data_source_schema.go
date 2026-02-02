// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TemplatesDataSourceSchema returns the schema for the mailgun_templates data source.
func TemplatesDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves a list of all templates for a Mailgun domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to list templates for.",
				Required:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "The total number of templates found.",
				Computed:    true,
			},
			"templates": schema.ListNestedAttribute{
				Description: "List of templates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the template.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the template.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The timestamp when the template was created.",
							Computed:    true,
						},
						"version_tag": schema.StringAttribute{
							Description: "The tag of the active template version.",
							Computed:    true,
						},
						"version_engine": schema.StringAttribute{
							Description: "The templating engine used (handlebars or go).",
							Computed:    true,
						},
						"version_active": schema.BoolAttribute{
							Description: "Whether the version is active.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
