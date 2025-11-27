// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ resource.Resource                = &ipAllowlistResource{}
	_ resource.ResourceWithConfigure   = &ipAllowlistResource{}
	_ resource.ResourceWithImportState = &ipAllowlistResource{}
)

// NewIPAllowlistResource creates a new IP allowlist resource.
func NewIPAllowlistResource() resource.Resource {
	return &ipAllowlistResource{}
}

type ipAllowlistResource struct {
	client *IPAllowlistClient
}

func (r *ipAllowlistResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_allowlist"
}

func (r *ipAllowlistResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = IPAllowlistResourceSchema()
}

func (r *ipAllowlistResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	mg, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = NewIPAllowlistClient(mg)
}

func (r *ipAllowlistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IPAllowlistModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := plan.Address.ValueString()
	description := plan.Description.ValueString()

	// Create the IP allowlist entry
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.CreateIPAllowlistEntry(createCtx, address, description)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating IP Allowlist Entry",
			fmt.Sprintf("Could not create IP allowlist entry for %s: %s", address, err.Error()),
		)
		return
	}

	// Set computed fields
	plan.Id = types.StringValue(address)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ipAllowlistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IPAllowlistModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := state.Address.ValueString()

	// Get the IP allowlist entry from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	entry, err := r.client.GetIPAllowlistEntry(readCtx, address)
	if err != nil {
		// Check if entry doesn't exist
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading IP Allowlist Entry",
			fmt.Sprintf("Could not read IP allowlist entry %s: %s", address, err.Error()),
		)
		return
	}

	// Update state with fetched data
	state.Address = types.StringValue(entry.IPAddress)
	state.Description = types.StringValue(entry.Description)
	state.Id = types.StringValue(entry.IPAddress)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ipAllowlistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IPAllowlistModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := plan.Address.ValueString()
	description := plan.Description.ValueString()

	// Update the IP allowlist entry
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.UpdateIPAllowlistEntry(updateCtx, address, description)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating IP Allowlist Entry",
			fmt.Sprintf("Could not update IP allowlist entry %s: %s", address, err.Error()),
		)
		return
	}

	// Set computed fields
	plan.Id = types.StringValue(address)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ipAllowlistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IPAllowlistModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := state.Address.ValueString()

	// Delete the IP allowlist entry
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteIPAllowlistEntry(deleteCtx, address)
	if err != nil {
		// Ignore not found errors during delete
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error Deleting IP Allowlist Entry",
				fmt.Sprintf("Could not delete IP allowlist entry %s: %s", address, err.Error()),
			)
			return
		}
	}
}

func (r *ipAllowlistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by IP address
	address := req.ID

	// Get the entry from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	entry, err := r.client.GetIPAllowlistEntry(importCtx, address)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing IP Allowlist Entry",
			fmt.Sprintf("Could not import IP allowlist entry %s: %s", address, err.Error()),
		)
		return
	}

	// Build state
	state := IPAllowlistModel{
		Address:     types.StringValue(entry.IPAddress),
		Description: types.StringValue(entry.Description),
		Id:          types.StringValue(entry.IPAddress),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
