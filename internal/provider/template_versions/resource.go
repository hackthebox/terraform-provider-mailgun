// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions

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
	_ resource.Resource                = &templateVersionResource{}
	_ resource.ResourceWithConfigure   = &templateVersionResource{}
	_ resource.ResourceWithImportState = &templateVersionResource{}
)

// NewTemplateVersionResource creates a new template version resource.
func NewTemplateVersionResource() resource.Resource {
	return &templateVersionResource{}
}

type templateVersionResource struct {
	client *mailgun.Client
}

func (r *templateVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_version"
}

func (r *templateVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TemplateVersionResourceSchema()
}

func (r *templateVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *templateVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TemplateVersionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	templateName := plan.TemplateName.ValueString()

	// Build version object
	version := &mtypes.TemplateVersion{
		Template: plan.Template.ValueString(),
		Engine:   mtypes.TemplateEngine(plan.Engine.ValueString()),
		Active:   plan.Active.ValueBool(),
	}

	if !plan.Tag.IsNull() && plan.Tag.ValueString() != "" {
		version.Tag = plan.Tag.ValueString()
	}

	if !plan.Comment.IsNull() {
		version.Comment = plan.Comment.ValueString()
	}

	// Create the version
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.AddTemplateVersion(createCtx, domain, templateName, version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Template Version",
			fmt.Sprintf("Could not create template version for %s/%s: %s", domain, templateName, err.Error()),
		)
		return
	}

	// If tag wasn't provided, we need to fetch the template to get the auto-generated tag
	// The API returns the created version in the template's Version field
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	template, err := r.client.GetTemplate(readCtx, domain, templateName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Template After Version Creation",
			fmt.Sprintf("Could not read template %s/%s: %s", domain, templateName, err.Error()),
		)
		return
	}

	// Use the version tag from the response or the one we set
	versionTag := version.Tag
	if versionTag == "" {
		versionTag = template.Version.Tag
	}

	// Read back the specific version to get computed values
	createdVersion, err := r.client.GetTemplateVersion(readCtx, domain, templateName, versionTag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Template Version",
			fmt.Sprintf("Could not read template version %s/%s/%s: %s", domain, templateName, versionTag, err.Error()),
		)
		return
	}

	// Map response to state
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", domain, templateName, createdVersion.Tag))
	plan.Tag = types.StringValue(createdVersion.Tag)
	plan.CreatedAt = types.StringValue(createdVersion.CreatedAt.String())
	plan.Active = types.BoolValue(createdVersion.Active)
	plan.Engine = types.StringValue(string(createdVersion.Engine))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *templateVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateVersionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	templateName := state.TemplateName.ValueString()
	tag := state.Tag.ValueString()

	// Get version from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	version, err := r.client.GetTemplateVersion(readCtx, domain, templateName, tag)
	if err != nil {
		// Check if version doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Template Version",
			fmt.Sprintf("Could not read template version %s/%s/%s: %s", domain, templateName, tag, err.Error()),
		)
		return
	}

	// Update state with fetched data
	state.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", domain, templateName, version.Tag))
	state.Tag = types.StringValue(version.Tag)
	state.Template = types.StringValue(version.Template)
	state.Engine = types.StringValue(string(version.Engine))
	state.Active = types.BoolValue(version.Active)
	state.CreatedAt = types.StringValue(version.CreatedAt.String())

	if version.Comment != "" {
		state.Comment = types.StringValue(version.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *templateVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TemplateVersionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	templateName := plan.TemplateName.ValueString()
	tag := plan.Tag.ValueString()

	// Build update version object
	version := &mtypes.TemplateVersion{
		Tag:      tag,
		Template: plan.Template.ValueString(),
		Engine:   mtypes.TemplateEngine(plan.Engine.ValueString()),
		Active:   plan.Active.ValueBool(),
	}

	if !plan.Comment.IsNull() {
		version.Comment = plan.Comment.ValueString()
	}

	// Update the version
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.UpdateTemplateVersion(updateCtx, domain, templateName, version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mailgun Template Version",
			fmt.Sprintf("Could not update template version %s/%s/%s: %s", domain, templateName, tag, err.Error()),
		)
		return
	}

	// Read back to get updated state
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	updatedVersion, err := r.client.GetTemplateVersion(readCtx, domain, templateName, tag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Template Version",
			fmt.Sprintf("Could not read template version %s/%s/%s: %s", domain, templateName, tag, err.Error()),
		)
		return
	}

	// Update state
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", domain, templateName, updatedVersion.Tag))
	plan.CreatedAt = types.StringValue(updatedVersion.CreatedAt.String())
	plan.Active = types.BoolValue(updatedVersion.Active)
	plan.Engine = types.StringValue(string(updatedVersion.Engine))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *templateVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TemplateVersionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	templateName := state.TemplateName.ValueString()
	tag := state.Tag.ValueString()

	// Delete the version
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteTemplateVersion(deleteCtx, domain, templateName, tag)
	if err != nil {
		errStr := err.Error()
		// Ignore not found errors during delete
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "404") {
			return
		}
		// Provide a clearer error message for active version deletion
		if strings.Contains(errStr, "deleting active version is not allowed") {
			resp.Diagnostics.AddError(
				"Cannot Delete Active Template Version",
				fmt.Sprintf("Template version %s/%s/%s is currently active and cannot be deleted. "+
					"To delete this version, first make another version active, or delete the parent template resource instead.",
					domain, templateName, tag),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Mailgun Template Version",
			fmt.Sprintf("Could not delete template version %s/%s/%s: %s", domain, templateName, tag, err.Error()),
		)
		return
	}
}

func (r *templateVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: domain/template_name/tag
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'domain/template_name/tag', got: %s", req.ID),
		)
		return
	}

	domain := parts[0]
	templateName := parts[1]
	tag := parts[2]

	// Get version from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	version, err := r.client.GetTemplateVersion(importCtx, domain, templateName, tag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Template Version",
			fmt.Sprintf("Could not import template version %s/%s/%s: %s", domain, templateName, tag, err.Error()),
		)
		return
	}

	// Build state
	state := TemplateVersionModel{
		Domain:       types.StringValue(domain),
		TemplateName: types.StringValue(templateName),
		Tag:          types.StringValue(version.Tag),
		Template:     types.StringValue(version.Template),
		Engine:       types.StringValue(string(version.Engine)),
		Active:       types.BoolValue(version.Active),
		Id:           types.StringValue(req.ID),
		CreatedAt:    types.StringValue(version.CreatedAt.String()),
	}

	if version.Comment != "" {
		state.Comment = types.StringValue(version.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
