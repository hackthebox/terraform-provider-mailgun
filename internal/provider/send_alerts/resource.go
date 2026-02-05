// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ resource.Resource                = &sendAlertResource{}
	_ resource.ResourceWithConfigure   = &sendAlertResource{}
	_ resource.ResourceWithImportState = &sendAlertResource{}
)

// NewSendAlertResource creates a new send alert resource.
func NewSendAlertResource() resource.Resource {
	return &sendAlertResource{}
}

type sendAlertResource struct {
	client *mailgun.Client
}

func (r *sendAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_send_alert"
}

func (r *sendAlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SendAlertResourceSchema()
}

func (r *sendAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sendAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SendAlertModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request
	apiReq, err := r.buildAPIRequest(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Building Request",
			fmt.Sprintf("Could not build send alert request: %s", err.Error()),
		)
		return
	}

	// Make API call
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(r.client)
	alertResp, err := apiClient.CreateSendAlert(createCtx, *apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Send Alert",
			fmt.Sprintf("Could not create send alert: %s", err.Error()),
		)
		return
	}

	// Map response to state
	r.mapAPIResponseToState(ctx, alertResp, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *sendAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SendAlertModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(r.client)
	alertResp, err := apiClient.GetSendAlert(readCtx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Send Alert",
			fmt.Sprintf("Could not read send alert %s: %s", name, err.Error()),
		)
		return
	}

	if alertResp == nil {
		// Alert not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to state
	r.mapAPIResponseToState(ctx, alertResp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *sendAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SendAlertModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SendAlertModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	// Build API request
	apiReq, err := r.buildAPIRequest(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Building Request",
			fmt.Sprintf("Could not build send alert request: %s", err.Error()),
		)
		return
	}

	// Make API call
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(r.client)
	err = apiClient.UpdateSendAlert(updateCtx, name, *apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mailgun Send Alert",
			fmt.Sprintf("Could not update send alert %s: %s", name, err.Error()),
		)
		return
	}

	// Read back the updated alert to get computed values
	alertResp, err := apiClient.GetSendAlert(updateCtx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Mailgun Send Alert",
			fmt.Sprintf("Could not read updated send alert: %s", err.Error()),
		)
		return
	}

	if alertResp != nil {
		r.mapAPIResponseToState(ctx, alertResp, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *sendAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SendAlertModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(r.client)
	err := apiClient.DeleteSendAlert(deleteCtx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Mailgun Send Alert",
			fmt.Sprintf("Could not delete send alert %s: %s", name, err.Error()),
		)
		return
	}
}

func (r *sendAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(r.client)
	alertResp, err := apiClient.GetSendAlert(readCtx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Send Alert",
			fmt.Sprintf("Could not import send alert %s: %s", name, err.Error()),
		)
		return
	}

	if alertResp == nil {
		resp.Diagnostics.AddError(
			"Send Alert Not Found",
			fmt.Sprintf("Send alert with name '%s' was not found.", name),
		)
		return
	}

	var state SendAlertModel
	r.mapAPIResponseToState(ctx, alertResp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// buildAPIRequest constructs the API request from the Terraform model.
func (r *sendAlertResource) buildAPIRequest(ctx context.Context, model *SendAlertModel) (*SendAlertAPIRequest, error) {
	req := &SendAlertAPIRequest{
		Name:       model.Name.ValueString(),
		Metric:     model.Metric.ValueString(),
		Comparator: model.Comparator.ValueString(),
		Limit:      model.Limit.ValueString(),
		Dimension:  model.Dimension.ValueString(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		req.Description = model.Description.ValueString()
	}

	if !model.Period.IsNull() && !model.Period.IsUnknown() {
		req.Period = model.Period.ValueString()
	}

	// Convert alert channels
	if !model.AlertChannels.IsNull() && !model.AlertChannels.IsUnknown() {
		var channels []string
		diags := model.AlertChannels.ElementsAs(ctx, &channels, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse alert_channels")
		}
		req.AlertChannels = channels
	}

	// Convert filters
	if !model.Filters.IsNull() && !model.Filters.IsUnknown() {
		var filters []SendAlertFilterModel
		diags := model.Filters.ElementsAs(ctx, &filters, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse filters")
		}

		for _, f := range filters {
			apiFilter := SendAlertFilterAPIObject{
				Dimension: f.Dimension.ValueString(),
			}
			if !f.Comparator.IsNull() && !f.Comparator.IsUnknown() {
				apiFilter.Comparator = f.Comparator.ValueString()
			}
			if !f.Values.IsNull() && !f.Values.IsUnknown() {
				var values []string
				diags := f.Values.ElementsAs(ctx, &values, false)
				if diags.HasError() {
					return nil, fmt.Errorf("failed to parse filter values")
				}
				apiFilter.Values = values
			}
			req.Filters = append(req.Filters, apiFilter)
		}
	}

	return req, nil
}

// mapAPIResponseToState maps the API response to the Terraform state model.
func (r *sendAlertResource) mapAPIResponseToState(ctx context.Context, apiResp *SendAlertAPIResponse, state *SendAlertModel, diagnostics *diag.Diagnostics) {

	state.ID = types.StringValue(apiResp.ID)
	state.Name = types.StringValue(apiResp.Name)
	state.Metric = types.StringValue(apiResp.Metric)
	state.Limit = types.StringValue(apiResp.Limit)
	state.Dimension = types.StringValue(apiResp.Dimension)
	state.CreatedAt = types.StringValue(apiResp.CreatedAt)

	// Optional/computed fields
	if apiResp.ParentAccountID != "" {
		state.ParentAccountID = types.StringValue(apiResp.ParentAccountID)
	} else {
		state.ParentAccountID = types.StringNull()
	}

	if apiResp.SubaccountID != "" {
		state.SubaccountID = types.StringValue(apiResp.SubaccountID)
	} else {
		state.SubaccountID = types.StringNull()
	}

	if apiResp.AccountGroup != "" {
		state.AccountGroup = types.StringValue(apiResp.AccountGroup)
	} else {
		state.AccountGroup = types.StringNull()
	}

	if apiResp.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(apiResp.UpdatedAt)
	} else {
		state.UpdatedAt = types.StringNull()
	}

	if apiResp.LastChecked != "" {
		state.LastChecked = types.StringValue(apiResp.LastChecked)
	} else {
		state.LastChecked = types.StringNull()
	}

	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}

	if apiResp.Period != "" {
		state.Period = types.StringValue(apiResp.Period)
	} else {
		state.Period = types.StringNull()
	}

	if apiResp.Comparator != "" {
		state.Comparator = types.StringValue(apiResp.Comparator)
	}

	// Alert channels
	if len(apiResp.AlertChannels) > 0 {
		channels, d := types.ListValueFrom(ctx, types.StringType, apiResp.AlertChannels)
		diagnostics.Append(d...)
		state.AlertChannels = channels
	} else {
		state.AlertChannels = types.ListNull(types.StringType)
	}

	// Filters
	if len(apiResp.Filters) > 0 {
		filterObjects := make([]attr.Value, 0, len(apiResp.Filters))
		for _, f := range apiResp.Filters {
			values, d := types.ListValueFrom(ctx, types.StringType, f.Values)
			diagnostics.Append(d...)

			filterObj, d := types.ObjectValue(
				FilterObjectType(),
				map[string]attr.Value{
					"dimension":  types.StringValue(f.Dimension),
					"comparator": types.StringValue(f.Comparator),
					"values":     values,
				},
			)
			diagnostics.Append(d...)
			filterObjects = append(filterObjects, filterObj)
		}
		filters, d := types.ListValue(types.ObjectType{AttrTypes: FilterObjectType()}, filterObjects)
		diagnostics.Append(d...)
		state.Filters = filters
	} else {
		state.Filters = types.ListNull(types.ObjectType{AttrTypes: FilterObjectType()})
	}
}
