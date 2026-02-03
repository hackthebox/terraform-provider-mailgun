// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TemplateVersionsDataSourceSchema returns the schema for the mailgun_template_versions data source.
func TemplateVersionsDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves a list of all versions for a Mailgun template.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name that owns the template.",
				Required:    true,
			},
			"template_name": schema.StringAttribute{
				Description: "The name of the template to list versions for.",
				Required:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "The total number of versions found.",
				Computed:    true,
			},
			"versions": schema.ListNestedAttribute{
				Description: "List of template versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tag": schema.StringAttribute{
							Description: "The version tag identifier.",
							Computed:    true,
						},
						"engine": schema.StringAttribute{
							Description: "The templating engine used (handlebars or go).",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "The comment for this version.",
							Computed:    true,
						},
						"active": schema.BoolAttribute{
							Description: "Whether this version is active.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The timestamp when the version was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
