// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// TemplateResourceSchema returns the schema for the mailgun_template resource.
func TemplateResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Mailgun template. Templates allow you to store and reuse email content with variable substitution using Handlebars or Go templating engines.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name that owns this template.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the template. Must be unique within the domain.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the template.",
				Optional:    true,
			},
			"template": schema.StringAttribute{
				Description: "The template content (HTML or text). This creates the initial version of the template.",
				Optional:    true,
			},
			"engine": schema.StringAttribute{
				Description: "The templating engine to use. Valid values: 'handlebars' (default) or 'go'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("handlebars"),
				Validators: []validator.String{
					stringvalidator.OneOf("handlebars", "go"),
				},
			},
			"comment": schema.StringAttribute{
				Description: "A comment for the initial template version.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the template (domain/name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the template was created.",
				Computed:    true,
			},
			"version_tag": schema.StringAttribute{
				Description: "The tag of the active template version.",
				Computed:    true,
			},
			"version_active": schema.BoolAttribute{
				Description: "Whether the current version is active.",
				Computed:    true,
			},
		},
	}
}
