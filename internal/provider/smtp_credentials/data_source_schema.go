// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// SmtpCredentialDataSourceSchema returns the schema for the single SMTP credential data source
func SmtpCredentialDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		Description: "Retrieves information about a specific SMTP credential for a Mailgun domain.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Description: "The unique identifier for this credential in format 'domain/login'.",
				Computed:    true,
			},
			"domain": dschema.StringAttribute{
				Description: "The domain the SMTP credential belongs to.",
				Required:    true,
			},
			"login": dschema.StringAttribute{
				Description: "The login name for SMTP authentication (without the @domain part).",
				Required:    true,
			},
			"full_login": dschema.StringAttribute{
				Description: "The full SMTP login in format 'login@domain'. Use this value for SMTP authentication.",
				Computed:    true,
			},
			"created_at": dschema.StringAttribute{
				Description: "The timestamp when this credential was created.",
				Computed:    true,
			},
		},
	}
}

// SmtpCredentialsListDataSourceSchema returns the schema for the SMTP credentials list data source
func SmtpCredentialsListDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		Description: "Lists all SMTP credentials for a Mailgun domain.",
		Attributes: map[string]dschema.Attribute{
			"domain": dschema.StringAttribute{
				Description: "The domain to list SMTP credentials for.",
				Required:    true,
			},
			"limit": dschema.Int64Attribute{
				Description: "Maximum number of credentials to return. Defaults to 100.",
				Optional:    true,
			},
			"total_count": dschema.Int64Attribute{
				Description: "The total number of credentials returned.",
				Computed:    true,
			},
			"credentials": dschema.ListNestedAttribute{
				Description: "List of SMTP credentials.",
				Computed:    true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"login": dschema.StringAttribute{
							Description: "The login name (without the @domain part).",
							Computed:    true,
						},
						"full_login": dschema.StringAttribute{
							Description: "The full SMTP login in format 'login@domain'.",
							Computed:    true,
						},
						"created_at": dschema.StringAttribute{
							Description: "The timestamp when this credential was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// NewSmtpCredentialDataSource creates a new single SMTP credential data source
func NewSmtpCredentialDataSource() datasource.DataSource {
	return &SmtpCredentialDataSource{}
}

// NewSmtpCredentialsListDataSource creates a new SMTP credentials list data source
func NewSmtpCredentialsListDataSource() datasource.DataSource {
	return &SmtpCredentialsListDataSource{}
}
