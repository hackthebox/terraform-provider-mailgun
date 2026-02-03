// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

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
	_ resource.Resource                = &mailingListMemberResource{}
	_ resource.ResourceWithConfigure   = &mailingListMemberResource{}
	_ resource.ResourceWithImportState = &mailingListMemberResource{}
)

// NewMailingListMemberResource creates a new mailing list member resource.
func NewMailingListMemberResource() resource.Resource {
	return &mailingListMemberResource{}
}

type mailingListMemberResource struct {
	client *mailgun.Client
}

func (r *mailingListMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailing_list_member"
}

func (r *mailingListMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MailingListMemberResourceSchema()
}

func (r *mailingListMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mailingListMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MailingListMemberModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listAddress := plan.ListAddress.ValueString()
	memberAddress := plan.MemberAddress.ValueString()

	// Build member object
	member := mtypes.Member{
		Address: memberAddress,
	}

	if !plan.Name.IsNull() {
		member.Name = plan.Name.ValueString()
	}

	subscribed := plan.Subscribed.ValueBool()
	member.Subscribed = &subscribed

	// Convert vars map
	if !plan.Vars.IsNull() {
		vars := make(map[string]any)
		diags = plan.Vars.ElementsAs(ctx, &vars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		member.Vars = vars
	}

	// Create the member
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.CreateMember(createCtx, false, listAddress, member)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Mailing List Member",
			fmt.Sprintf("Could not create member %s in list %s: %s", memberAddress, listAddress, err.Error()),
		)
		return
	}

	// Read back to verify
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	createdMember, err := r.client.GetMember(readCtx, memberAddress, listAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Member",
			fmt.Sprintf("Could not read member %s after creation: %s", memberAddress, err.Error()),
		)
		return
	}

	// Map response to state
	mapMemberToModel(&createdMember, &plan, ctx)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *mailingListMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MailingListMemberModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listAddress := state.ListAddress.ValueString()
	memberAddress := state.MemberAddress.ValueString()

	// Get member from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	member, err := r.client.GetMember(readCtx, memberAddress, listAddress)
	if err != nil {
		// Check if member doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Mailing List Member",
			fmt.Sprintf("Could not read member %s from list %s: %s", memberAddress, listAddress, err.Error()),
		)
		return
	}

	// Map response to state
	mapMemberToModel(&member, &state, ctx)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *mailingListMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MailingListMemberModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listAddress := plan.ListAddress.ValueString()
	memberAddress := plan.MemberAddress.ValueString()

	// Build member object
	member := mtypes.Member{
		Address: memberAddress,
	}

	if !plan.Name.IsNull() {
		member.Name = plan.Name.ValueString()
	}

	subscribed := plan.Subscribed.ValueBool()
	member.Subscribed = &subscribed

	// Convert vars map
	if !plan.Vars.IsNull() {
		vars := make(map[string]any)
		diags = plan.Vars.ElementsAs(ctx, &vars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		member.Vars = vars
	}

	// Update the member
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := r.client.UpdateMember(updateCtx, memberAddress, listAddress, member)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mailgun Mailing List Member",
			fmt.Sprintf("Could not update member %s in list %s: %s", memberAddress, listAddress, err.Error()),
		)
		return
	}

	// Read back to get updated state
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	updatedMember, err := r.client.GetMember(readCtx, memberAddress, listAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Member",
			fmt.Sprintf("Could not read member %s after update: %s", memberAddress, err.Error()),
		)
		return
	}

	// Map response to state
	mapMemberToModel(&updatedMember, &plan, ctx)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *mailingListMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MailingListMemberModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listAddress := state.ListAddress.ValueString()
	memberAddress := state.MemberAddress.ValueString()

	// Delete the member
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteMember(deleteCtx, memberAddress, listAddress)
	if err != nil {
		// Ignore not found errors during delete
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error Deleting Mailgun Mailing List Member",
				fmt.Sprintf("Could not delete member %s from list %s: %s", memberAddress, listAddress, err.Error()),
			)
			return
		}
	}
}

func (r *mailingListMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: list_address/member_address
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'list_address/member_address', got: %s", req.ID),
		)
		return
	}

	listAddress := parts[0]
	memberAddress := parts[1]

	// Get member from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	member, err := r.client.GetMember(importCtx, memberAddress, listAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Mailing List Member",
			fmt.Sprintf("Could not import member %s from list %s: %s", memberAddress, listAddress, err.Error()),
		)
		return
	}

	// Build state
	var state MailingListMemberModel
	state.ListAddress = types.StringValue(listAddress)
	state.MemberAddress = types.StringValue(memberAddress)
	mapMemberToModel(&member, &state, ctx)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// mapMemberToModel maps a Mailgun member to the Terraform model
func mapMemberToModel(member *mtypes.Member, model *MailingListMemberModel, ctx context.Context) {
	model.Id = types.StringValue(fmt.Sprintf("%s/%s", model.ListAddress.ValueString(), member.Address))
	model.MemberAddress = types.StringValue(member.Address)

	if member.Name != "" {
		model.Name = types.StringValue(member.Name)
	} else {
		model.Name = types.StringNull()
	}

	if member.Subscribed != nil {
		model.Subscribed = types.BoolValue(*member.Subscribed)
	} else {
		model.Subscribed = types.BoolValue(true)
	}

	// Convert vars to map
	if len(member.Vars) > 0 {
		stringVars := make(map[string]string)
		for k, v := range member.Vars {
			if str, ok := v.(string); ok {
				stringVars[k] = str
			} else {
				stringVars[k] = fmt.Sprintf("%v", v)
			}
		}
		varsMap, _ := types.MapValueFrom(ctx, types.StringType, stringVars)
		model.Vars = varsMap
	} else {
		model.Vars = types.MapNull(types.StringType)
	}
}
