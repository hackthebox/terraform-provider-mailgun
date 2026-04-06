// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

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
	_ resource.Resource                = &domainSendingKeyResource{}
	_ resource.ResourceWithConfigure   = &domainSendingKeyResource{}
	_ resource.ResourceWithImportState = &domainSendingKeyResource{}
)

// NewDomainSendingKeyResource creates a new domain sending key resource.
func NewDomainSendingKeyResource() resource.Resource {
	return &domainSendingKeyResource{}
}

type domainSendingKeyResource struct {
	client *mailgun.Client
}

func (r *domainSendingKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_sending_key"
}

func (r *domainSendingKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainSendingKeyResourceSchema()
}

func (r *domainSendingKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainSendingKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainSendingKeyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	// Build create options - hardcode kind=domain for domain sending keys
	opts := &mailgun.CreateAPIKeyOptions{
		Kind:       "domain",
		DomainName: domain,
	}

	if !plan.Description.IsNull() && plan.Description.ValueString() != "" {
		opts.Description = plan.Description.ValueString()
	}

	if !plan.Expiration.IsNull() && plan.Expiration.ValueInt64() > 0 {
		opts.Expiration = uint64(plan.Expiration.ValueInt64())
	}

	// Create the API key with role=sending (hardcoded for domain sending keys)
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiKey, err := r.client.CreateAPIKey(createCtx, "sending", opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Domain Sending Key",
			fmt.Sprintf("Could not create sending key for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Map response to state - secret is only available now!
	plan.Id = types.StringValue(apiKey.ID)
	plan.Secret = types.StringValue(apiKey.Secret)
	plan.CreatedAt = types.StringValue(apiKey.CreatedAt.Format(time.RFC3339))

	if !apiKey.ExpiresAt.IsZero() {
		plan.ExpiresAt = types.StringValue(apiKey.ExpiresAt.Format(time.RFC3339))
	} else {
		plan.ExpiresAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainSendingKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainSendingKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.Id.ValueString()
	domain := state.Domain.ValueString()

	// Find the API key
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiKey, err := r.findKey(readCtx, keyID, domain)
	if err != nil {
		// Key not found - remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with API data, but preserve the secret
	// (secret is not returned by ListAPIKeys)
	secret := state.Secret
	state.Id = types.StringValue(apiKey.ID)
	state.Domain = types.StringValue(apiKey.DomainName)
	state.Description = types.StringValue(apiKey.Description)
	state.CreatedAt = types.StringValue(apiKey.CreatedAt.Format(time.RFC3339))
	state.Secret = secret // Preserve the secret from state

	if !apiKey.ExpiresAt.IsZero() {
		state.ExpiresAt = types.StringValue(apiKey.ExpiresAt.Format(time.RFC3339))
	} else {
		state.ExpiresAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *domainSendingKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API keys are immutable - all changes require replacement via plan modifiers
	resp.Diagnostics.AddError(
		"Domain Sending Keys Cannot Be Updated",
		"Mailgun API keys are immutable after creation. All changes require creating a new key.",
	)
}

func (r *domainSendingKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainSendingKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.Id.ValueString()

	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteAPIKey(deleteCtx, keyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Domain Sending Key",
			fmt.Sprintf("Could not delete sending key %s: %s", keyID, err.Error()),
		)
		return
	}
}

func (r *domainSendingKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	keyID := req.ID

	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Search for the key across all domain keys
	apiKey, err := r.findKeyByID(importCtx, keyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Domain Sending Key",
			fmt.Sprintf("Could not find sending key %s: %s", keyID, err.Error()),
		)
		return
	}

	// Verify it's a domain sending key
	if apiKey.Kind != "domain" || apiKey.Role != "sending" {
		resp.Diagnostics.AddError(
			"Invalid Key Type",
			fmt.Sprintf("Key %s is not a domain sending key (kind=%s, role=%s). "+
				"This resource only manages domain sending keys.", keyID, apiKey.Kind, apiKey.Role),
		)
		return
	}

	// Build state from imported data
	state := DomainSendingKeyModel{
		Id:          types.StringValue(apiKey.ID),
		Domain:      types.StringValue(apiKey.DomainName),
		Description: types.StringValue(apiKey.Description),
		Secret:      types.StringNull(), // Secret cannot be retrieved after creation
		CreatedAt:   types.StringValue(apiKey.CreatedAt.Format(time.RFC3339)),
	}

	if !apiKey.ExpiresAt.IsZero() {
		state.ExpiresAt = types.StringValue(apiKey.ExpiresAt.Format(time.RFC3339))
	} else {
		state.ExpiresAt = types.StringNull()
	}

	// Default expiration to 0 (we can't retrieve the original value)
	state.Expiration = types.Int64Value(0)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.AddWarning(
		"API Key Secret Not Available",
		"The sending key was imported successfully, but the secret cannot be retrieved. "+
			"The secret was only available immediately after creation. "+
			"If you need the secret, create a new key.",
	)
}

// findKey searches for a domain sending key by ID
func (r *domainSendingKeyResource) findKey(ctx context.Context, keyID, domain string) (*mtypes.APIKey, error) {
	opts := &mailgun.ListAPIKeysOptions{
		DomainName: domain,
		Kind:       "domain",
	}

	keys, err := r.client.ListAPIKeys(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error listing API keys: %w", err)
	}

	for _, key := range keys {
		if key.ID == keyID {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("sending key not found")
}

// findKeyByID searches for a domain sending key by ID across all domains
func (r *domainSendingKeyResource) findKeyByID(ctx context.Context, keyID string) (*mtypes.APIKey, error) {
	// List all domain keys
	opts := &mailgun.ListAPIKeysOptions{
		Kind: "domain",
	}

	keys, err := r.client.ListAPIKeys(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error listing API keys: %w", err)
	}

	for _, key := range keys {
		if key.ID == keyID {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("sending key %s not found", keyID)
}
