// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &sendAlertDataSource{}
	_ datasource.DataSourceWithConfigure = &sendAlertDataSource{}
)

// NewSendAlertDataSource creates a new send alert data source.
func NewSendAlertDataSource() datasource.DataSource {
	return &sendAlertDataSource{}
}

type sendAlertDataSource struct {
	client *mailgun.Client
}

func (d *sendAlertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_send_alert"
}

func (d *sendAlertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SendAlertDataSourceSchema()
}

func (d *sendAlertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *sendAlertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SendAlertModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(d.client)
	alertResp, err := apiClient.GetSendAlert(readCtx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Send Alert",
			fmt.Sprintf("Could not read send alert %s: %s", name, err.Error()),
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

	// Map response to state
	d.mapAPIResponseToState(ctx, alertResp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// mapAPIResponseToState maps the API response to the Terraform state model.
func (d *sendAlertDataSource) mapAPIResponseToState(ctx context.Context, apiResp *SendAlertAPIResponse, state *SendAlertModel, diagnostics *diag.Diagnostics) {

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
	} else {
		state.Comparator = types.StringNull()
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
