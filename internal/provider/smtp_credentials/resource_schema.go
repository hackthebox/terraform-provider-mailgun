// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// SmtpCredentialResourceSchema returns the schema for the SMTP credential resource
func SmtpCredentialResourceSchema() rschema.Schema {
	return rschema.Schema{
		Version:     0,
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
				Description: "The password for SMTP authentication. This is write-only and cannot be read back from the API. " +
					"Set this when creating a credential or when rotating the password of an imported credential. " +
					"Leave it unset to keep the existing password of an imported credential.",
				DeprecationMessage: "Use password_wo and password_wo_version instead. The password argument stores the " +
					"secret in Terraform state; it remains supported for backward compatibility but will be removed in a future major release.",
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password_wo": rschema.StringAttribute{
				Description: "Write-only password for SMTP authentication. The value is never stored in Terraform state. " +
					"Set it together with password_wo_version and increment the version to rotate the password. Requires Terraform CLI >= 1.11.",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("password")),
					stringvalidator.AlsoRequires(path.MatchRoot("password_wo_version")),
				},
			},
			"password_wo_version": rschema.Int64Attribute{
				Description: "Version counter for password_wo. Increment this value to rotate the write-only password. Required when password_wo is set.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("password_wo")),
					int64validator.ConflictsWith(path.MatchRoot("password")),
				},
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
