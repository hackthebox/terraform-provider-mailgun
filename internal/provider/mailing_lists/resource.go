// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists

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
	_ resource.Resource                = &mailingListResource{}
	_ resource.ResourceWithConfigure   = &mailingListResource{}
	_ resource.ResourceWithImportState = &mailingListResource{}
)

// NewMailingListResource creates a new mailing list resource.
func NewMailingListResource() resource.Resource {
	return &mailingListResource{}
}

type mailingListResource struct {
	client *mailgun.Client
}

func (r *mailingListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailing_list"
}

func (r *mailingListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MailingListResourceSchema()
}

func (r *mailingListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mailingListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MailingListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := plan.Address.ValueString()

	// Build mailing list object
	list := mtypes.MailingList{
		Address:     address,
		AccessLevel: mtypes.AccessLevel(plan.AccessLevel.ValueString()),
	}

	if !plan.Name.IsNull() {
		list.Name = plan.Name.ValueString()
	}

	if !plan.Description.IsNull() {
		list.Description = plan.Description.ValueString()
	}

	if !plan.ReplyPreference.IsNull() {
		list.ReplyPreference = mtypes.ReplyPreference(plan.ReplyPreference.ValueString())
	}

	// Create the mailing list
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := r.client.CreateMailingList(createCtx, list)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Mailing List",
			fmt.Sprintf("Could not create mailing list %s: %s", address, err.Error()),
		)
		return
	}

	// Read back to get computed values
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	createdList, err := r.client.GetMailingList(readCtx, address)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Mailing List",
			fmt.Sprintf("Could not read mailing list %s after creation: %s", address, err.Error()),
		)
		return
	}

	// Map response to state
	mapMailingListToModel(&createdList, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *mailingListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MailingListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := state.Address.ValueString()

	// Get mailing list from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	list, err := r.client.GetMailingList(readCtx, address)
	if err != nil {
		// Check if list doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Mailing List",
			fmt.Sprintf("Could not read mailing list %s: %s", address, err.Error()),
		)
		return
	}

	// Map response to state
	mapMailingListToModel(&list, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *mailingListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MailingListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := plan.Address.ValueString()

	// Build update object
	list := mtypes.MailingList{
		Address:     address,
		AccessLevel: mtypes.AccessLevel(plan.AccessLevel.ValueString()),
	}

	if !plan.Name.IsNull() {
		list.Name = plan.Name.ValueString()
	}

	if !plan.Description.IsNull() {
		list.Description = plan.Description.ValueString()
	}

	if !plan.ReplyPreference.IsNull() {
		list.ReplyPreference = mtypes.ReplyPreference(plan.ReplyPreference.ValueString())
	}

	// Update the mailing list
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := r.client.UpdateMailingList(updateCtx, address, list)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mailgun Mailing List",
			fmt.Sprintf("Could not update mailing list %s: %s", address, err.Error()),
		)
		return
	}

	// Read back to get updated state
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	updatedList, err := r.client.GetMailingList(readCtx, address)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Mailing List",
			fmt.Sprintf("Could not read mailing list %s after update: %s", address, err.Error()),
		)
		return
	}

	// Map response to state
	mapMailingListToModel(&updatedList, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *mailingListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MailingListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := state.Address.ValueString()

	// Delete the mailing list
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteMailingList(deleteCtx, address)
	if err != nil {
		// Ignore not found errors during delete
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error Deleting Mailgun Mailing List",
				fmt.Sprintf("Could not delete mailing list %s: %s", address, err.Error()),
			)
			return
		}
	}
}

func (r *mailingListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by mailing list address
	address := req.ID

	// Get mailing list from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	list, err := r.client.GetMailingList(importCtx, address)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Mailing List",
			fmt.Sprintf("Could not import mailing list %s: %s", address, err.Error()),
		)
		return
	}

	// Build state
	var state MailingListModel
	state.Address = types.StringValue(address)
	mapMailingListToModel(&list, &state)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// mapMailingListToModel maps a Mailgun mailing list to the Terraform model
func mapMailingListToModel(list *mtypes.MailingList, model *MailingListModel) {
	model.Id = types.StringValue(list.Address)
	model.Address = types.StringValue(list.Address)
	model.CreatedAt = types.StringValue(list.CreatedAt.String())
	model.MembersCount = types.Int64Value(int64(list.MembersCount))
	model.AccessLevel = types.StringValue(string(list.AccessLevel))

	if list.Name != "" {
		model.Name = types.StringValue(list.Name)
	} else {
		model.Name = types.StringNull()
	}

	if list.Description != "" {
		model.Description = types.StringValue(list.Description)
	} else {
		model.Description = types.StringNull()
	}

	if list.ReplyPreference != "" {
		model.ReplyPreference = types.StringValue(string(list.ReplyPreference))
	} else {
		model.ReplyPreference = types.StringValue("list")
	}
}
