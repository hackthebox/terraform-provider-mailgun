// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ApiKeyResource{}
	_ resource.ResourceWithConfigure   = &ApiKeyResource{}
	_ resource.ResourceWithImportState = &ApiKeyResource{}
)

// ApiKeyResource is the resource implementation.
type ApiKeyResource struct {
	client *mailgun.Client
}

// Metadata returns the resource type name.
func (r *ApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// Schema defines the schema for the resource.
func (r *ApiKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ApiKeyResourceSchema()
}

// Configure adds the provider-configured client to the resource.
func (r *ApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new API key.
func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. Please report this issue to the provider developers.",
		)
		return
	}

	role := plan.Role.ValueString()
	if role == "" {
		resp.Diagnostics.AddError("Missing Role", "The role is required to create an API key.")
		return
	}

	// Build create options
	opts := &mailgun.CreateAPIKeyOptions{}

	if !plan.Description.IsNull() && plan.Description.ValueString() != "" {
		opts.Description = plan.Description.ValueString()
	}

	if !plan.DomainName.IsNull() && plan.DomainName.ValueString() != "" {
		opts.DomainName = plan.DomainName.ValueString()
	}

	if !plan.Kind.IsNull() && plan.Kind.ValueString() != "" {
		opts.Kind = plan.Kind.ValueString()
	}

	if !plan.Expiration.IsNull() && plan.Expiration.ValueInt64() > 0 {
		opts.Expiration = uint64(plan.Expiration.ValueInt64())
	}

	// Create context with timeout
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create the API key via Mailgun API
	apiKey, err := r.client.CreateAPIKey(createCtx, role, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating API Key",
			fmt.Sprintf("Could not create API key with role '%s': %s", role, err),
		)
		return
	}

	// Map response to state - this is critical as secret is only available now!
	plan = mapApiKeyToModel(apiKey, plan)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. Please report this issue to the provider developers.",
		)
		return
	}

	keyID := state.Id.ValueString()
	if keyID == "" {
		resp.Diagnostics.AddError("Missing API Key ID", "The API key ID is required to read the key.")
		return
	}

	// Find the API key
	apiKey, err := r.findApiKey(ctx, keyID, state.DomainName.ValueString(), state.Kind.ValueString())
	if err != nil {
		// API key not found - remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with API data, but preserve the secret from state
	// (secret is not returned by ListAPIKeys)
	secret := state.Secret
	state = mapApiKeyToModel(*apiKey, state)
	state.Secret = secret // Preserve the secret

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for API keys - they are immutable.
// All changes require replacement via plan modifiers.
func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This should not be called because all attributes have RequiresReplace plan modifier
	resp.Diagnostics.AddError(
		"API Keys Cannot Be Updated",
		"Mailgun API keys are immutable after creation. All changes require creating a new key.",
	)
}

// Delete deletes an existing API key.
func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiKeyModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. Please report this issue to the provider developers.",
		)
		return
	}

	keyID := state.Id.ValueString()
	if keyID == "" {
		resp.Diagnostics.AddError("Missing API Key ID", "The API key ID is required to delete the key.")
		return
	}

	// Create context with timeout
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Delete the API key via Mailgun API
	err := r.client.DeleteAPIKey(deleteCtx, keyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting API Key",
			fmt.Sprintf("Could not delete API key %s: %s", keyID, err),
		)
		return
	}

	// State is automatically removed by Terraform after successful deletion
}

// ImportState imports an existing API key by ID
func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	keyID := req.ID

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. Please report this issue to the provider developers.",
		)
		return
	}

	// Find the API key (search all types since we don't know the domain/kind)
	apiKey, err := r.findApiKeyByID(ctx, keyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing API Key",
			fmt.Sprintf("Could not find API key %s: %s", keyID, err),
		)
		return
	}

	// Create state from imported data
	var state ApiKeyModel
	state = mapApiKeyToModel(*apiKey, state)
	// Secret cannot be imported - it was only available at creation
	state.Secret = types.StringValue("")

	// Save imported state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	// Warn user that secret is not available
	resp.Diagnostics.AddWarning(
		"API Key Secret Not Available",
		"The API key was imported successfully, but the secret cannot be retrieved from the API. "+
			"The secret was only available immediately after the key was created. "+
			"If you need the secret, you must create a new API key.",
	)
}

// findApiKey searches for a specific API key by ID
func (r *ApiKeyResource) findApiKey(ctx context.Context, keyID, domainName, kind string) (*mtypes.APIKey, error) {
	// Create context with timeout
	findCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build list options
	opts := &mailgun.ListAPIKeysOptions{}
	if domainName != "" {
		opts.DomainName = domainName
	}
	if kind != "" {
		opts.Kind = kind
	}

	// List API keys
	keys, err := r.client.ListAPIKeys(findCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("error listing API keys: %w", err)
	}

	for _, key := range keys {
		if key.ID == keyID {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("API key not found")
}

// findApiKeyByID searches for an API key by ID across all types
func (r *ApiKeyResource) findApiKeyByID(ctx context.Context, keyID string) (*mtypes.APIKey, error) {
	// Try with no filters first
	key, err := r.findApiKey(ctx, keyID, "", "")
	if err == nil {
		return key, nil
	}

	// Try different kinds if not found
	kinds := []string{"domain", "user", "web"}
	for _, kind := range kinds {
		key, err := r.findApiKey(ctx, keyID, "", kind)
		if err == nil {
			return key, nil
		}
	}

	return nil, fmt.Errorf("API key %s not found", keyID)
}

// mapApiKeyToModel maps Mailgun API key response to Terraform model
func mapApiKeyToModel(apiKey mtypes.APIKey, model ApiKeyModel) ApiKeyModel {
	model.Id = types.StringValue(apiKey.ID)
	model.Role = types.StringValue(apiKey.Role)
	model.Kind = types.StringValue(apiKey.Kind)
	model.Description = types.StringValue(apiKey.Description)

	if apiKey.DomainName != "" {
		model.DomainName = types.StringValue(apiKey.DomainName)
	} else {
		model.DomainName = types.StringNull()
	}

	// Secret is only populated on creation
	if apiKey.Secret != "" {
		model.Secret = types.StringValue(apiKey.Secret)
	}
	// If Secret is empty and model already has a value, preserve it (for Read operations)

	model.CreatedAt = types.StringValue(apiKey.CreatedAt.Format(time.RFC3339))
	model.UpdatedAt = types.StringValue(apiKey.UpdatedAt.Format(time.RFC3339))

	if !apiKey.ExpiresAt.IsZero() {
		model.ExpiresAt = types.StringValue(apiKey.ExpiresAt.Format(time.RFC3339))
	} else {
		model.ExpiresAt = types.StringNull()
	}

	model.IsDisabled = types.BoolValue(apiKey.IsDisabled)
	model.DisabledReason = types.StringValue(apiKey.DisabledReason)
	model.Requestor = types.StringValue(apiKey.Requestor)
	model.UserName = types.StringValue(apiKey.UserName)

	return model
}
