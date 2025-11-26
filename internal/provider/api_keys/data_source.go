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
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ApiKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &ApiKeyDataSource{}
)

// ApiKeyDataSource is the single API key data source implementation.
type ApiKeyDataSource struct {
	client *mailgun.Client
}

// Metadata returns the data source type name.
func (d *ApiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// Schema defines the schema for the data source.
func (d *ApiKeyDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ApiKeyDataSourceSchema()
}

// Configure adds the provider-configured client to the data source.
func (d *ApiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ApiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApiKeyDataSourceModel

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

	keyID := data.Id.ValueString()
	if keyID == "" {
		resp.Diagnostics.AddError("Missing API Key ID", "The id is required to lookup an API key.")
		return
	}

	// Find the API key
	apiKey, err := d.findApiKeyByID(ctx, keyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading API Key",
			fmt.Sprintf("Could not find API key %s: %s", keyID, err),
		)
		return
	}

	// Map response to data source model
	data.Role = types.StringValue(apiKey.Role)
	data.Kind = types.StringValue(apiKey.Kind)
	data.Description = types.StringValue(apiKey.Description)

	if apiKey.DomainName != "" {
		data.DomainName = types.StringValue(apiKey.DomainName)
	} else {
		data.DomainName = types.StringNull()
	}

	data.CreatedAt = types.StringValue(apiKey.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(apiKey.UpdatedAt.Format(time.RFC3339))

	if !apiKey.ExpiresAt.IsZero() {
		data.ExpiresAt = types.StringValue(apiKey.ExpiresAt.Format(time.RFC3339))
	} else {
		data.ExpiresAt = types.StringNull()
	}

	data.IsDisabled = types.BoolValue(apiKey.IsDisabled)
	data.DisabledReason = types.StringValue(apiKey.DisabledReason)
	data.Requestor = types.StringValue(apiKey.Requestor)
	data.UserName = types.StringValue(apiKey.UserName)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// findApiKeyByID searches for an API key by ID across all types
func (d *ApiKeyDataSource) findApiKeyByID(ctx context.Context, keyID string) (*mtypes.APIKey, error) {
	// Create context with timeout
	findCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Try with no filters first
	keys, err := d.client.ListAPIKeys(findCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing API keys: %w", err)
	}

	for _, key := range keys {
		if key.ID == keyID {
			return &key, nil
		}
	}

	// Try different kinds if not found
	kinds := []string{"domain", "user", "web"}
	for _, kind := range kinds {
		keys, err := d.client.ListAPIKeys(findCtx, &mailgun.ListAPIKeysOptions{Kind: kind})
		if err != nil {
			continue
		}
		for _, key := range keys {
			if key.ID == keyID {
				return &key, nil
			}
		}
	}

	return nil, fmt.Errorf("API key %s not found", keyID)
}
