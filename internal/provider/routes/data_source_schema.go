// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoutesDataSourceSchema returns the schema for the mailgun_routes data source.
func RoutesDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Lists all Mailgun routes in your account.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Description: "Maximum number of routes to return. Defaults to 100.",
				Optional:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of routes returned.",
				Computed:    true,
			},
			"routes": schema.ListNestedAttribute{
				Description: "List of routes.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the route.",
							Computed:    true,
						},
						"expression": schema.StringAttribute{
							Description: "The filter expression for the route.",
							Computed:    true,
						},
						"actions": schema.ListAttribute{
							Description: "List of actions to execute when the expression matches.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"priority": schema.Int64Attribute{
							Description: "Route priority. Lower numbers have higher priority.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Human-readable description of the route.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the route was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
