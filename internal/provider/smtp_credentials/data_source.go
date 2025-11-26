// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &SmtpCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &SmtpCredentialDataSource{}
)

// SmtpCredentialDataSource is the single SMTP credential data source implementation.
type SmtpCredentialDataSource struct {
	client *mailgun.Client
}

// Metadata returns the data source type name.
func (d *SmtpCredentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_credential"
}

// Schema defines the schema for the data source.
func (d *SmtpCredentialDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SmtpCredentialDataSourceSchema()
}

// Configure adds the provider-configured client to the data source.
func (d *SmtpCredentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SmtpCredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SmtpCredentialDataSourceModel

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

	domain := data.Domain.ValueString()
	login := data.Login.ValueString()

	if domain == "" {
		resp.Diagnostics.AddError("Missing Domain", "The domain is required to lookup an SMTP credential.")
		return
	}
	if login == "" {
		resp.Diagnostics.AddError("Missing Login", "The login is required to lookup an SMTP credential.")
		return
	}

	// Find the credential
	credential, err := d.findCredential(ctx, domain, login)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SMTP Credential",
			fmt.Sprintf("Could not find SMTP credential %s@%s: %s", login, domain, err),
		)
		return
	}

	// Extract login part from full login if needed
	loginPart := login
	if strings.Contains(credential.Login, "@") {
		parts := strings.SplitN(credential.Login, "@", 2)
		loginPart = parts[0]
	}

	// Map response to data source model
	data.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, loginPart))
	data.FullLogin = types.StringValue(credential.Login)
	data.CreatedAt = types.StringValue(time.Time(credential.CreatedAt).Format(time.RFC3339))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// findCredential searches for a specific credential by domain and login
func (d *SmtpCredentialDataSource) findCredential(ctx context.Context, domain, login string) (*mtypes.Credential, error) {
	// Create context with timeout
	findCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List all credentials for the domain
	iterator := d.client.ListCredentials(domain, &mailgun.ListOptions{Limit: 100})

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
