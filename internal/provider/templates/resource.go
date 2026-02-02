// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ resource.Resource                = &templateResource{}
	_ resource.ResourceWithConfigure   = &templateResource{}
	_ resource.ResourceWithImportState = &templateResource{}
)

// NewTemplateResource creates a new template resource.
func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

type templateResource struct {
	client *mailgun.Client
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TemplateResourceSchema()
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TemplateModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	name := plan.Name.ValueString()

	// Build template object
	template := &mtypes.Template{
		Name: name,
	}

	if !plan.Description.IsNull() {
		template.Description = plan.Description.ValueString()
	}

	// If template content is provided, create initial version
	if !plan.Template.IsNull() && plan.Template.ValueString() != "" {
		template.Version = mtypes.TemplateVersion{
			Template: plan.Template.ValueString(),
			Engine:   mtypes.TemplateEngine(plan.Engine.ValueString()),
			Active:   true,
		}
		if !plan.Comment.IsNull() {
			template.Version.Comment = plan.Comment.ValueString()
		}
	}

	// Create the template
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.CreateTemplate(createCtx, domain, template)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Template",
			fmt.Sprintf("Could not create template %s for domain %s: %s", name, domain, err.Error()),
		)
		return
	}

	// Read back the template to get computed values
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	createdTemplate, err := r.client.GetTemplate(readCtx, domain, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Template",
			fmt.Sprintf("Could not read template %s after creation: %s", name, err.Error()),
		)
		return
	}

	// Map response to state
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, name))
	plan.CreatedAt = types.StringValue(createdTemplate.CreatedAt.String())

	if createdTemplate.Version.Tag != "" {
		plan.VersionTag = types.StringValue(createdTemplate.Version.Tag)
		plan.VersionActive = types.BoolValue(createdTemplate.Version.Active)
	} else {
		plan.VersionTag = types.StringNull()
		plan.VersionActive = types.BoolNull()
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	name := state.Name.ValueString()

	// Get template from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	template, err := r.client.GetTemplate(readCtx, domain, name)
	if err != nil {
		// Check if template doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Template",
			fmt.Sprintf("Could not read template %s for domain %s: %s", name, domain, err.Error()),
		)
		return
	}

	// Update state with fetched data
	state.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, name))
	state.Name = types.StringValue(template.Name)
	state.Description = types.StringValue(template.Description)
	state.CreatedAt = types.StringValue(template.CreatedAt.String())

	if template.Version.Tag != "" {
		state.VersionTag = types.StringValue(template.Version.Tag)
		state.VersionActive = types.BoolValue(template.Version.Active)
		state.Engine = types.StringValue(string(template.Version.Engine))
		if template.Version.Template != "" {
			state.Template = types.StringValue(template.Version.Template)
		}
		if template.Version.Comment != "" {
			state.Comment = types.StringValue(template.Version.Comment)
		}
	} else {
		state.VersionTag = types.StringNull()
		state.VersionActive = types.BoolNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TemplateModel
	var state TemplateModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	name := plan.Name.ValueString()

	// Update template description if changed
	if !plan.Description.Equal(state.Description) {
		template := &mtypes.Template{
			Name:        name,
			Description: plan.Description.ValueString(),
		}

		updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		err := r.client.UpdateTemplate(updateCtx, domain, template)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Mailgun Template",
				fmt.Sprintf("Could not update template %s for domain %s: %s", name, domain, err.Error()),
			)
			return
		}
	}

	// If template content changed, add a new version
	if !plan.Template.Equal(state.Template) && !plan.Template.IsNull() && plan.Template.ValueString() != "" {
		version := &mtypes.TemplateVersion{
			Template: plan.Template.ValueString(),
			Engine:   mtypes.TemplateEngine(plan.Engine.ValueString()),
			Active:   true,
		}
		if !plan.Comment.IsNull() {
			version.Comment = plan.Comment.ValueString()
		}

		versionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		err := r.client.AddTemplateVersion(versionCtx, domain, name, version)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Adding Template Version",
				fmt.Sprintf("Could not add version to template %s: %s", name, err.Error()),
			)
			return
		}
	}

	// Read back to get updated state
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	template, err := r.client.GetTemplate(readCtx, domain, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Template",
			fmt.Sprintf("Could not read template %s after update: %s", name, err.Error()),
		)
		return
	}

	// Update state
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, name))
	plan.CreatedAt = types.StringValue(template.CreatedAt.String())

	if template.Version.Tag != "" {
		plan.VersionTag = types.StringValue(template.Version.Tag)
		plan.VersionActive = types.BoolValue(template.Version.Active)
	} else {
		plan.VersionTag = types.StringNull()
		plan.VersionActive = types.BoolNull()
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TemplateModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	name := state.Name.ValueString()

	// Delete the template
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteTemplate(deleteCtx, domain, name)
	if err != nil {
		// Ignore not found errors during delete
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error Deleting Mailgun Template",
				fmt.Sprintf("Could not delete template %s for domain %s: %s", name, domain, err.Error()),
			)
			return
		}
	}
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: domain/template_name
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'domain/template_name', got: %s", req.ID),
		)
		return
	}

	domain := parts[0]
	name := parts[1]

	// Get template from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	template, err := r.client.GetTemplate(importCtx, domain, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Template",
			fmt.Sprintf("Could not import template %s for domain %s: %s", name, domain, err.Error()),
		)
		return
	}

	// Build state
	state := TemplateModel{
		Domain:      types.StringValue(domain),
		Name:        types.StringValue(template.Name),
		Description: types.StringValue(template.Description),
		Id:          types.StringValue(req.ID),
		CreatedAt:   types.StringValue(template.CreatedAt.String()),
	}

	if template.Version.Tag != "" {
		state.VersionTag = types.StringValue(template.Version.Tag)
		state.VersionActive = types.BoolValue(template.Version.Active)
		state.Engine = types.StringValue(string(template.Version.Engine))
		if template.Version.Template != "" {
			state.Template = types.StringValue(template.Version.Template)
		}
		if template.Version.Comment != "" {
			state.Comment = types.StringValue(template.Version.Comment)
		}
	} else {
		state.VersionTag = types.StringNull()
		state.VersionActive = types.BoolNull()
		state.Engine = types.StringValue("handlebars")
		state.Template = types.StringNull()
		state.Comment = types.StringNull()
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
