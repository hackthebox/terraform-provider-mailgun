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
	_ datasource.DataSource              = &sendAlertsListDataSource{}
	_ datasource.DataSourceWithConfigure = &sendAlertsListDataSource{}
)

// NewSendAlertsListDataSource creates a new send alerts list data source.
func NewSendAlertsListDataSource() datasource.DataSource {
	return &sendAlertsListDataSource{}
}

type sendAlertsListDataSource struct {
	client *mailgun.Client
}

func (d *sendAlertsListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_send_alerts"
}

func (d *sendAlertsListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SendAlertsListDataSourceSchema()
}

func (d *sendAlertsListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sendAlertsListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SendAlertsListModel

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiClient := NewSendAlertsAPIClient(d.client)
	alertsResp, err := apiClient.ListSendAlerts(readCtx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Send Alerts",
			fmt.Sprintf("Could not list send alerts: %s", err.Error()),
		)
		return
	}

	state.TotalCount = types.Int64Value(int64(alertsResp.Total))

	// Build list of alerts
	alertObjects := make([]attr.Value, 0, len(alertsResp.Items))
	for _, alert := range alertsResp.Items {
		alertObj, diags := d.mapAlertToObject(ctx, &alert)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		alertObjects = append(alertObjects, alertObj)
	}

	alertsList, diags := types.ListValue(d.alertObjectType(), alertObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Alerts = alertsList

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// alertObjectType returns the object type for a single alert in the list.
func (d *sendAlertsListDataSource) alertObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":                types.StringType,
			"parent_account_id": types.StringType,
			"subaccount_id":     types.StringType,
			"account_group":     types.StringType,
			"created_at":        types.StringType,
			"updated_at":        types.StringType,
			"last_checked":      types.StringType,
			"name":              types.StringType,
			"metric":            types.StringType,
			"comparator":        types.StringType,
			"limit":             types.StringType,
			"dimension":         types.StringType,
			"description":       types.StringType,
			"period":            types.StringType,
			"alert_channels":    types.ListType{ElemType: types.StringType},
			"filters":           types.ListType{ElemType: types.ObjectType{AttrTypes: FilterObjectType()}},
		},
	}
}

// mapAlertToObject maps an API alert response to a Terraform object value.
func (d *sendAlertsListDataSource) mapAlertToObject(ctx context.Context, alert *SendAlertAPIResponse) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Alert channels
	var alertChannels types.List
	if len(alert.AlertChannels) > 0 {
		var d diag.Diagnostics
		alertChannels, d = types.ListValueFrom(ctx, types.StringType, alert.AlertChannels)
		diags.Append(d...)
	} else {
		alertChannels = types.ListNull(types.StringType)
	}

	// Filters
	var filters types.List
	if len(alert.Filters) > 0 {
		filterObjects := make([]attr.Value, 0, len(alert.Filters))
		for _, f := range alert.Filters {
			values, d := types.ListValueFrom(ctx, types.StringType, f.Values)
			diags.Append(d...)

			filterObj, d := types.ObjectValue(
				FilterObjectType(),
				map[string]attr.Value{
					"dimension":  types.StringValue(f.Dimension),
					"comparator": types.StringValue(f.Comparator),
					"values":     values,
				},
			)
			diags.Append(d...)
			filterObjects = append(filterObjects, filterObj)
		}
		var d diag.Diagnostics
		filters, d = types.ListValue(types.ObjectType{AttrTypes: FilterObjectType()}, filterObjects)
		diags.Append(d...)
	} else {
		filters = types.ListNull(types.ObjectType{AttrTypes: FilterObjectType()})
	}

	// Build object
	obj, objDiags := types.ObjectValue(
		d.alertObjectType().AttrTypes,
		map[string]attr.Value{
			"id":                stringOrNull(alert.ID),
			"parent_account_id": stringOrNull(alert.ParentAccountID),
			"subaccount_id":     stringOrNull(alert.SubaccountID),
			"account_group":     stringOrNull(alert.AccountGroup),
			"created_at":        types.StringValue(alert.CreatedAt),
			"updated_at":        stringOrNull(alert.UpdatedAt),
			"last_checked":      stringOrNull(alert.LastChecked),
			"name":              types.StringValue(alert.Name),
			"metric":            types.StringValue(alert.Metric),
			"comparator":        stringOrNull(alert.Comparator),
			"limit":             types.StringValue(alert.Limit),
			"dimension":         types.StringValue(alert.Dimension),
			"description":       stringOrNull(alert.Description),
			"period":            stringOrNull(alert.Period),
			"alert_channels":    alertChannels,
			"filters":           filters,
		},
	)
	diags.Append(objDiags...)

	return obj, diags
}

// stringOrNull returns a types.String with the value or null if empty.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
