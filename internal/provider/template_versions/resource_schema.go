// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// TemplateVersionResourceSchema returns the schema for the mailgun_template_version resource.
func TemplateVersionResourceSchema() schema.Schema {
	return schema.Schema{
		Version:     0,
		Description: "Manages a version of a Mailgun template. Template versions allow you to maintain multiple versions of email content for A/B testing or rollbacks.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name that owns the template.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_name": schema.StringAttribute{
				Description: "The name of the parent template.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template": schema.StringAttribute{
				Description: "The template content (HTML or text).",
				Required:    true,
			},
			"tag": schema.StringAttribute{
				Description: "The version tag identifier. If not specified, Mailgun auto-generates one. Changing this forces recreation.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Description: "A comment describing this version.",
				Optional:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether this version is the active version of the template. The first version created is automatically active. Note: Active versions cannot be deleted directly; delete the parent template instead.",
				Optional:    true,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the template version (domain/template_name/tag).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the version was created.",
				Computed:    true,
			},
		},
	}
}
