// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &SmtpCredentialResource{}
	_ resource.ResourceWithConfigure   = &SmtpCredentialResource{}
	_ resource.ResourceWithImportState = &SmtpCredentialResource{}
)

// SmtpCredentialResource is the resource implementation.
type SmtpCredentialResource struct {
	client *mailgun.Client
}

// Metadata returns the resource type name.
func (r *SmtpCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_credential"
}

// Schema defines the schema for the resource.
func (r *SmtpCredentialResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SmtpCredentialResourceSchema()
}

// Configure adds the provider-configured client to the resource.
func (r *SmtpCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new SMTP credential.
func (r *SmtpCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SmtpCredentialModel

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

	domain := plan.Domain.ValueString()
	login := plan.Login.ValueString()
	password := plan.Password.ValueString()

	// Validate required fields
	if domain == "" {
		resp.Diagnostics.AddError("Missing Domain", "The domain is required to create an SMTP credential.")
		return
	}
	if login == "" {
		resp.Diagnostics.AddError("Missing Login", "The login is required to create an SMTP credential.")
		return
	}
	if plan.Password.IsNull() || plan.Password.IsUnknown() || password == "" {
		resp.Diagnostics.AddError("Missing Password", "The password is required to create an SMTP credential.")
		return
	}

	// Create context with timeout
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create the credential via Mailgun API
	err := r.client.CreateCredential(createCtx, domain, login, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SMTP Credential",
			fmt.Sprintf("Could not create SMTP credential %s@%s: %s", login, domain, err),
		)
		return
	}

	// Set computed values
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, login))
	plan.FullLogin = types.StringValue(fmt.Sprintf("%s@%s", login, domain))

	// Try to get the created_at from the API by listing credentials
	credential, err := r.findCredential(ctx, domain, login)
	if err != nil {
		// Not fatal - we created it, just can't get the timestamp
		plan.CreatedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))
	} else {
		plan.CreatedAt = types.StringValue(time.Time(credential.CreatedAt).Format(time.RFC3339))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SmtpCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SmtpCredentialModel

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

	domain := state.Domain.ValueString()
	login := state.Login.ValueString()

	// Find the credential in the API
	credential, err := r.findCredential(ctx, domain, login)
	if err != nil {
		// Credential not found - remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with API data
	state.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, login))
	state.FullLogin = types.StringValue(credential.Login)
	state.CreatedAt = types.StringValue(time.Time(credential.CreatedAt).Format(time.RFC3339))
	// Note: Password is not returned by the API, keep the value from state

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing SMTP credential (only password can be changed).
func (r *SmtpCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SmtpCredentialModel
	var state SmtpCredentialModel

	// Read Terraform plan and state data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	domain := plan.Domain.ValueString()
	login := plan.Login.ValueString()
	password := plan.Password.ValueString()

	// Password is write-only and cannot be read back from the API. When it is
	// omitted from configuration for an imported resource, preserve the
	// existing state and do not attempt a password rotation.
	if plan.Password.IsNull() {
		plan.Password = state.Password
		plan.Id = state.Id
		plan.FullLogin = state.FullLogin
		plan.CreatedAt = state.CreatedAt

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Invalid Password",
			"The password cannot be an empty string. Omit the attribute to preserve the existing imported password, or set a non-empty value to rotate it.",
		)
		return
	}

	// Create context with timeout
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Only the password can be updated
	err := r.client.ChangeCredentialPassword(updateCtx, domain, login, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating SMTP Credential",
			fmt.Sprintf("Could not update password for SMTP credential %s@%s: %s", login, domain, err),
		)
		return
	}

	// Keep computed values from state, update password from plan
	plan.Id = state.Id
	plan.FullLogin = state.FullLogin
	plan.CreatedAt = state.CreatedAt

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes an existing SMTP credential.
func (r *SmtpCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SmtpCredentialModel

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

	domain := state.Domain.ValueString()
	login := state.Login.ValueString()

	// Create context with timeout
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Delete the credential via Mailgun API
	err := r.client.DeleteCredential(deleteCtx, domain, login)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SMTP Credential",
			fmt.Sprintf("Could not delete SMTP credential %s@%s: %s", login, domain, err),
		)
		return
	}

	// State is automatically removed by Terraform after successful deletion
}

// ImportState imports an existing SMTP credential by ID (format: domain/login)
func (r *SmtpCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the ID (format: domain/login)
	idParts := strings.SplitN(req.ID, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'domain/login', got: %s", req.ID),
		)
		return
	}

	domain := idParts[0]
	login := idParts[1]

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. Please report this issue to the provider developers.",
		)
		return
	}

	// Find the credential in the API
	credential, err := r.findCredential(ctx, domain, login)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing SMTP Credential",
			fmt.Sprintf("Could not find SMTP credential %s@%s: %s", login, domain, err),
		)
		return
	}

	// Create state from imported data
	state := SmtpCredentialModel{
		Id:        types.StringValue(fmt.Sprintf("%s/%s", domain, login)),
		Domain:    types.StringValue(domain),
		Login:     types.StringValue(login),
		FullLogin: types.StringValue(credential.Login),
		CreatedAt: types.StringValue(time.Time(credential.CreatedAt).Format(time.RFC3339)),
		// Password cannot be imported because the API never returns it.
		Password: types.StringNull(),
	}

	// Save imported state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// findCredential searches for a specific credential by domain and login
func (r *SmtpCredentialResource) findCredential(ctx context.Context, domain, login string) (*mtypes.Credential, error) {
	// Create context with timeout
	findCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List all credentials for the domain
	iterator := r.client.ListCredentials(domain, &mailgun.ListOptions{Limit: 100})

	var page []mtypes.Credential
	for iterator.Next(findCtx, &page) {
		for _, cred := range page {
			// The API returns full login (login@domain), so we need to compare appropriately
			expectedFullLogin := fmt.Sprintf("%s@%s", login, domain)
			if cred.Login == expectedFullLogin || cred.Login == login {
				return &cred, nil
			}
		}
	}

	if err := iterator.Err(); err != nil {
		return nil, fmt.Errorf("error listing credentials: %w", err)
	}

	return nil, fmt.Errorf("credential not found")
}
