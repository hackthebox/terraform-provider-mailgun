// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RouteResourceSchema returns the schema for the mailgun_route resource.
func RouteResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Mailgun route. Routes allow you to handle incoming emails by matching expressions and executing actions.",
		Attributes: map[string]schema.Attribute{
			"expression": schema.StringAttribute{
				Description: "A filter expression like 'match_recipient(\".*@example.com\")' or 'match_header(\"subject\", \".*support.*\")'. " +
					"See https://documentation.mailgun.com/docs/mailgun/user-manual/routes/ for expression syntax.",
				Required: true,
			},
			"actions": schema.ListAttribute{
				Description: "List of actions to execute when the expression matches. " +
					"Examples: 'forward(\"http://example.com/webhook\")', 'store(notify=\"http://example.com\")', 'stop()'.",
				Required:    true,
				ElementType: types.StringType,
			},
			"priority": schema.Int64Attribute{
				Description: "Route priority. Lower numbers have higher priority. Routes with equal priority are evaluated in chronological order. Defaults to 0.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the route.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the route.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the route was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
