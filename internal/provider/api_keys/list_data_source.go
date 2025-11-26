// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ApiKeysListDataSource{}
	_ datasource.DataSourceWithConfigure = &ApiKeysListDataSource{}
)

// ApiKeysListDataSource is the API keys list data source implementation.
type ApiKeysListDataSource struct {
	client *mailgun.Client
}

// Metadata returns the data source type name.
func (d *ApiKeysListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

// Schema defines the schema for the data source.
func (d *ApiKeysListDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ApiKeysListDataSourceSchema()
}

// Configure adds the provider-configured client to the data source.
func (d *ApiKeysListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *ApiKeysListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApiKeysListDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. Please report this issue to the provider developers.",
		)
		return
	}

	// Build list options
	opts := &mailgun.ListAPIKeysOptions{}
	if !data.DomainName.IsNull() && data.DomainName.ValueString() != "" {
		opts.DomainName = data.DomainName.ValueString()
	}
	if !data.Kind.IsNull() && data.Kind.ValueString() != "" {
		opts.Kind = data.Kind.ValueString()
	}

	// Create context with timeout
	listCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List API keys from Mailgun API
	keys, err := d.client.ListAPIKeys(listCtx, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing API Keys",
			fmt.Sprintf("Could not list API keys: %s", err),
		)
		return
	}

	// Convert to Terraform model
	keyItems := make([]ApiKeyItemModel, len(keys))
	for i, key := range keys {
		keyItems[i] = ApiKeyItemModel{
			Id:          types.StringValue(key.ID),
			Role:        types.StringValue(key.Role),
			Kind:        types.StringValue(key.Kind),
			Description: types.StringValue(key.Description),
			CreatedAt:   types.StringValue(key.CreatedAt.Format(time.RFC3339)),
			UpdatedAt:   types.StringValue(key.UpdatedAt.Format(time.RFC3339)),
			IsDisabled:  types.BoolValue(key.IsDisabled),
			Requestor:   types.StringValue(key.Requestor),
			UserName:    types.StringValue(key.UserName),
		}

		if key.DomainName != "" {
			keyItems[i].DomainName = types.StringValue(key.DomainName)
		} else {
			keyItems[i].DomainName = types.StringNull()
		}

		if !key.ExpiresAt.IsZero() {
			keyItems[i].ExpiresAt = types.StringValue(key.ExpiresAt.Format(time.RFC3339))
		} else {
			keyItems[i].ExpiresAt = types.StringNull()
		}

		keyItems[i].DisabledReason = types.StringValue(key.DisabledReason)
	}

	// Set values in the data model
	data.Keys = keyItems
	data.TotalCount = types.Int64Value(int64(len(keys)))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
