// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_dkim_key"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_ip"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_sending_keys"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_tracking"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domains"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/ip_allowlist"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mailing_list_members"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mailing_lists"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/routes"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/smtp_credentials"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/template_versions"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/templates"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/webhooks"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &mailgunProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mailgunProvider{
			version: version,
		}
	}
}

// mailgunProvider is the provider implementation.
type mailgunProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// mailgunProviderModel describes the provider configuration.
type mailgunProviderModel struct {
	ApiKey   types.String `tfsdk:"api_key"`
	Region   types.String `tfsdk:"region"`
	Endpoint types.String `tfsdk:"endpoint"`
}

// Metadata returns the provider type name.
func (p *mailgunProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mailgun"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *mailgunProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Mailgun resources",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Mailgun API key for authentication. Can also be set via MAILGUN_API_KEY environment variable.",
				Required:    true,
				Sensitive:   true,
			},
			"region": schema.StringAttribute{
				Description: "Mailgun API region. Valid values: 'US' (default) or 'EU'.",
				Optional:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Custom Mailgun API endpoint. Overrides region setting.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a Mailgun API client for data sources and resources.
func (p *mailgunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config mailgunProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiKey.IsNull() {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"While configuring the provider, the API key was not found in "+
				"the Mailgun provider configuration. "+
				"Please set the api_key value in the provider configuration.",
		)
		return
	}

	// Default values
	endpoint := "https://api.mailgun.net"
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	// Create Mailgun client (v5 API)
	mg := mailgun.NewMailgun(config.ApiKey.ValueString())

	// Set region if provided
	if !config.Region.IsNull() {
		region := config.Region.ValueString()
		if region == "EU" {
			if err := mg.SetAPIBase("https://api.eu.mailgun.net/v3"); err != nil {
				resp.Diagnostics.AddError(
					"Failed to Set API Base for EU Region",
					"Could not configure the Mailgun client for EU region: "+err.Error(),
				)
				return
			}
		}
	}

	// If endpoint is provided, override the API base
	if !config.Endpoint.IsNull() {
		if err := mg.SetAPIBase(endpoint); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Set Custom API Endpoint",
				"Could not configure the Mailgun client with custom endpoint: "+err.Error(),
			)
			return
		}
	}

	// Make the Mailgun client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = mg
	resp.ResourceData = mg
}

// DataSources defines the data sources implemented in the provider.
func (p *mailgunProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		domains.NewDomainDataSource,                          // Single domain lookup by name
		domains.NewDomainsListDataSource,                     // List all domains
		smtp_credentials.NewSmtpCredentialDataSource,         // Single SMTP credential lookup
		smtp_credentials.NewSmtpCredentialsListDataSource,    // List SMTP credentials for a domain
		domain_sending_keys.NewDomainSendingKeysDataSource,   // List domain sending keys
		routes.NewRoutesDataSource,                           // List routes
		webhooks.NewWebhooksDataSource,                       // List webhooks for a domain
		ip_allowlist.NewIPAllowlistDataSource,                // List IP allowlist entries
		templates.NewTemplatesDataSource,                     // List templates for a domain
		template_versions.NewTemplateVersionsDataSource,      // List template versions
		mailing_lists.NewMailingListsDataSource,              // List mailing lists
		mailing_list_members.NewMailingListMembersDataSource, // List mailing list members
		domain_tracking.NewDomainTrackingDataSource,          // Domain tracking settings
		domain_dkim_key.NewDomainDkimKeysDataSource,          // List DKIM keys for a domain
		domain_ip.NewDomainIPsDataSource,                     // List IPs for a domain
	}
}

// Resources defines the resources implemented in the provider.
func (p *mailgunProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		domains.NewDomainResource,
		smtp_credentials.NewSmtpCredentialResource,
		domain_sending_keys.NewDomainSendingKeyResource,
		routes.NewRouteResource,
		webhooks.NewWebhookResource,
		ip_allowlist.NewIPAllowlistResource,
		templates.NewTemplateResource,
		template_versions.NewTemplateVersionResource,
		mailing_lists.NewMailingListResource,
		mailing_list_members.NewMailingListMemberResource,
		domain_tracking.NewDomainTrackingResource,
		domain_dkim_key.NewDomainDkimKeyResource,
		domain_ip.NewDomainIPResource,
	}
}
