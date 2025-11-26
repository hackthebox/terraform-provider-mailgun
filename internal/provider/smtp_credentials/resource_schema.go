// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// SmtpCredentialResourceSchema returns the schema for the SMTP credential resource
func SmtpCredentialResourceSchema() rschema.Schema {
	return rschema.Schema{
		Description: "Manages an SMTP credential for a Mailgun domain. SMTP credentials allow sending email via SMTP protocol.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Description: "The unique identifier for this credential in format 'domain/login'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": rschema.StringAttribute{
				Description: "The domain this SMTP credential belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"login": rschema.StringAttribute{
				Description: "The login name for SMTP authentication (without the @domain part). The full SMTP username will be 'login@domain'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": rschema.StringAttribute{
				Description: "The password for SMTP authentication. This is write-only and cannot be read back from the API.",
				Required:    true,
				Sensitive:   true,
			},
			"full_login": rschema.StringAttribute{
				Description: "The full SMTP login in format 'login@domain'. Use this value for SMTP authentication.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": rschema.StringAttribute{
				Description: "The timestamp when this credential was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// NewSmtpCredentialResource creates a new SMTP credential resource
func NewSmtpCredentialResource() resource.Resource {
	return &SmtpCredentialResource{}
}
