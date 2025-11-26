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
	_ datasource.DataSource              = &SmtpCredentialsListDataSource{}
	_ datasource.DataSourceWithConfigure = &SmtpCredentialsListDataSource{}
)

// SmtpCredentialsListDataSource is the SMTP credentials list data source implementation.
type SmtpCredentialsListDataSource struct {
	client *mailgun.Client
}

// Metadata returns the data source type name.
func (d *SmtpCredentialsListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_credentials"
}

// Schema defines the schema for the data source.
func (d *SmtpCredentialsListDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SmtpCredentialsListDataSourceSchema()
}

// Configure adds the provider-configured client to the data source.
func (d *SmtpCredentialsListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SmtpCredentialsListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SmtpCredentialsListDataSourceModel

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
	if domain == "" {
		resp.Diagnostics.AddError("Missing Domain", "The domain is required to list SMTP credentials.")
		return
	}

	// Set default limit
	limit := int64(100)
	if !data.Limit.IsNull() {
		limit = data.Limit.ValueInt64()
	}

	// Create context with timeout
	listCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List credentials from Mailgun API
	iterator := d.client.ListCredentials(domain, &mailgun.ListOptions{Limit: int(limit)})

	var credentials []mtypes.Credential
	var page []mtypes.Credential
	for iterator.Next(listCtx, &page) {
		credentials = append(credentials, page...)
	}

	if err := iterator.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Listing SMTP Credentials",
			fmt.Sprintf("Could not list SMTP credentials for domain %s: %s", domain, err),
		)
		return
	}

	// Convert to Terraform model
	credentialItems := make([]SmtpCredentialItemModel, len(credentials))
	for i, cred := range credentials {
		// Extract login part from full login
		loginPart := cred.Login
		if strings.Contains(cred.Login, "@") {
			parts := strings.SplitN(cred.Login, "@", 2)
			loginPart = parts[0]
		}

		credentialItems[i] = SmtpCredentialItemModel{
			Login:     types.StringValue(loginPart),
			FullLogin: types.StringValue(cred.Login),
			CreatedAt: types.StringValue(time.Time(cred.CreatedAt).Format(time.RFC3339)),
		}
	}

	// Set values in the data model
	data.Credentials = credentialItems
	data.TotalCount = types.Int64Value(int64(len(credentials)))
	if data.Limit.IsNull() {
		data.Limit = types.Int64Value(limit)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
