// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ datasource.DataSource              = &routesDataSource{}
	_ datasource.DataSourceWithConfigure = &routesDataSource{}
)

// NewRoutesDataSource creates a new routes data source.
func NewRoutesDataSource() datasource.DataSource {
	return &routesDataSource{}
}

type routesDataSource struct {
	client *mailgun.Client
}

func (d *routesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routes"
}

func (d *routesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RoutesDataSourceSchema()
}

func (d *routesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *routesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RoutesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default limit
	limit := int64(100)
	if !config.Limit.IsNull() {
		limit = config.Limit.ValueInt64()
	}

	// List routes
	opts := &mailgun.ListOptions{
		Limit: int(limit),
	}

	iter := d.client.ListRoutes(opts)

	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var routes []mtypes.Route
	var page []mtypes.Route
	for iter.Next(readCtx, &page) {
		routes = append(routes, page...)
		if len(routes) >= int(limit) {
			break
		}
	}

	if iter.Err() != nil {
		resp.Diagnostics.AddError(
			"Error Listing Mailgun Routes",
			fmt.Sprintf("Could not list routes: %s", iter.Err().Error()),
		)
		return
	}

	// Map routes to model
	routeItems := make([]attr.Value, len(routes))
	for i, route := range routes {
		actionsList, _ := types.ListValueFrom(ctx, types.StringType, route.Actions)

		routeObj, objDiags := types.ObjectValue(
			map[string]attr.Type{
				"id":          types.StringType,
				"expression":  types.StringType,
				"actions":     types.ListType{ElemType: types.StringType},
				"priority":    types.Int64Type,
				"description": types.StringType,
				"created_at":  types.StringType,
			},
			map[string]attr.Value{
				"id":          types.StringValue(route.Id),
				"expression":  types.StringValue(route.Expression),
				"actions":     actionsList,
				"priority":    types.Int64Value(int64(route.Priority)),
				"description": types.StringValue(route.Description),
				"created_at":  types.StringValue(time.Time(route.CreatedAt).Format(time.RFC3339)),
			},
		)
		resp.Diagnostics.Append(objDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		routeItems[i] = routeObj
	}

	routesList, listDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":          types.StringType,
				"expression":  types.StringType,
				"actions":     types.ListType{ElemType: types.StringType},
				"priority":    types.Int64Type,
				"description": types.StringType,
				"created_at":  types.StringType,
			},
		},
		routeItems,
	)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Routes = routesList
	config.TotalCount = types.Int64Value(int64(len(routes)))
	if config.Limit.IsNull() {
		config.Limit = types.Int64Value(limit)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
