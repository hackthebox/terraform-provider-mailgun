// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ datasource.DataSource              = &webhooksDataSource{}
	_ datasource.DataSourceWithConfigure = &webhooksDataSource{}
)

// NewWebhooksDataSource creates a new webhooks data source.
func NewWebhooksDataSource() datasource.DataSource {
	return &webhooksDataSource{}
}

type webhooksDataSource struct {
	client *mailgun.Client
}

func (d *webhooksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhooks"
}

func (d *webhooksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = WebhooksDataSourceSchema()
}

func (d *webhooksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *webhooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WebhooksDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := config.Domain.ValueString()

	// List webhooks for the domain
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	webhooksMap, err := d.client.ListWebhooks(readCtx, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Mailgun Webhooks",
			fmt.Sprintf("Could not list webhooks for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Sort webhook types for consistent ordering
	webhookTypes := make([]string, 0, len(webhooksMap))
	for webhookType := range webhooksMap {
		webhookTypes = append(webhookTypes, webhookType)
	}
	sort.Strings(webhookTypes)

	// Map webhooks to model
	webhookItems := make([]attr.Value, 0, len(webhooksMap))
	for _, webhookType := range webhookTypes {
		urls := webhooksMap[webhookType]
		urlsList, _ := types.ListValueFrom(ctx, types.StringType, urls)

		webhookObj, objDiags := types.ObjectValue(
			map[string]attr.Type{
				"webhook_type": types.StringType,
				"urls":         types.ListType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"webhook_type": types.StringValue(webhookType),
				"urls":         urlsList,
			},
		)
		resp.Diagnostics.Append(objDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		webhookItems = append(webhookItems, webhookObj)
	}

	webhooksList, listDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"webhook_type": types.StringType,
				"urls":         types.ListType{ElemType: types.StringType},
			},
		},
		webhookItems,
	)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Webhooks = webhooksList
	config.TotalCount = types.Int64Value(int64(len(webhooksMap)))

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
