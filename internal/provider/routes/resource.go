// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ resource.Resource                = &routeResource{}
	_ resource.ResourceWithConfigure   = &routeResource{}
	_ resource.ResourceWithImportState = &routeResource{}
)

// NewRouteResource creates a new route resource.
func NewRouteResource() resource.Resource {
	return &routeResource{}
}

type routeResource struct {
	client *mailgun.Client
}

func (r *routeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *routeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = RouteResourceSchema()
}

func (r *routeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RouteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert actions list to []string
	var actions []string
	diags = plan.Actions.ElementsAs(ctx, &actions, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build route request
	route := mtypes.Route{
		Expression:  plan.Expression.ValueString(),
		Actions:     actions,
		Priority:    int(plan.Priority.ValueInt64()),
		Description: plan.Description.ValueString(),
	}

	// Create the route
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	createdRoute, err := r.client.CreateRoute(createCtx, route)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Route",
			fmt.Sprintf("Could not create route: %s", err.Error()),
		)
		return
	}

	// Map response to state
	mapRouteToModel(&createdRoute, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RouteModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get route from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	route, err := r.client.GetRoute(readCtx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Route",
			fmt.Sprintf("Could not read route %s: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	// Map response to state
	mapRouteToModel(&route, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RouteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RouteModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert actions list to []string
	var actions []string
	diags = plan.Actions.ElementsAs(ctx, &actions, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build route update request
	route := mtypes.Route{
		Expression:  plan.Expression.ValueString(),
		Actions:     actions,
		Priority:    int(plan.Priority.ValueInt64()),
		Description: plan.Description.ValueString(),
	}

	// Update the route
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	updatedRoute, err := r.client.UpdateRoute(updateCtx, state.Id.ValueString(), route)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mailgun Route",
			fmt.Sprintf("Could not update route %s: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	// Map response to state
	mapRouteToModel(&updatedRoute, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RouteModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the route
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteRoute(deleteCtx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Mailgun Route",
			fmt.Sprintf("Could not delete route %s: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}
}

func (r *routeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by route ID
	routeId := req.ID

	// Get route from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	route, err := r.client.GetRoute(importCtx, routeId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Route",
			fmt.Sprintf("Could not import route %s: %s", routeId, err.Error()),
		)
		return
	}

	// Map response to state
	var state RouteModel
	mapRouteToModel(&route, &state)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// mapRouteToModel maps a Mailgun Route to a Terraform RouteModel.
func mapRouteToModel(route *mtypes.Route, model *RouteModel) {
	model.Id = types.StringValue(route.Id)
	model.Expression = types.StringValue(route.Expression)
	model.Priority = types.Int64Value(int64(route.Priority))
	model.Description = types.StringValue(route.Description)
	model.CreatedAt = types.StringValue(time.Time(route.CreatedAt).Format(time.RFC3339))

	// Convert actions to list
	actionValues := make([]types.String, len(route.Actions))
	for i, action := range route.Actions {
		actionValues[i] = types.StringValue(action)
	}
	model.Actions, _ = types.ListValueFrom(context.Background(), types.StringType, route.Actions)
}
