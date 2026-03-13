// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainResourceSchema returns the schema for the domain resource
func DomainResourceSchema() rschema.Schema {
	return rschema.Schema{
		Version:     0,
		Description: "Manages a Mailgun domain. Domains are used to send and receive email.",
		Attributes: map[string]rschema.Attribute{
			// Required/Optional input attributes
			"name": rschema.StringAttribute{
				Description: "The domain name to be used for sending and receiving email.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"smtp_password": rschema.StringAttribute{
				Description: "Password for SMTP authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"spam_action": rschema.StringAttribute{
				Description: "Spam filter action for new domain. Options: 'disabled', 'tag', or 'delete'. Changing this forces recreation.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "tag", "delete"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // SDK doesn't support updating spam_action
				},
			},
			"wildcard": rschema.BoolAttribute{
				Description: "Determines whether the domain will accept email for sub-domains. Changing this forces recreation.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(), // SDK doesn't support updating wildcard
				},
			},
			"force_dkim_authority": rschema.BoolAttribute{
				Description: "If set to true, the domain will be the DKIM authority for itself even if the root domain is registered on the same mailgun account. Changing this forces recreation.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(), // SDK doesn't support updating force_dkim_authority
				},
			},
			"dkim_key_size": rschema.StringAttribute{
				Description: "DKIM key size (1024 or 2048). Changing this forces recreation.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("1024", "2048"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // SDK doesn't support updating dkim_key_size
				},
			},
			"ips": rschema.StringAttribute{
				Description: "Comma-separated list of IPs to be assigned to this domain.",
				Optional:    true,
			},
			"pool_id": rschema.StringAttribute{
				Description: "The id of the IP Pool that you wish to assign to the domain.",
				Optional:    true,
			},
			"web_scheme": rschema.StringAttribute{
				Description: "Web scheme for tracking links (http or https).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("http", "https"),
				},
			},
			"web_prefix": rschema.StringAttribute{
				Description: "Web prefix for tracking links.",
				Optional:    true,
				Computed:    true,
			},
			"use_automatic_sender_security": rschema.BoolAttribute{
				Description: "Whether to use automatic sender security (SPF/DKIM/DMARC).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dkim_selector": rschema.StringAttribute{
				Description: "DKIM selector (custom CNAME for DKIM).",
				Optional:    true,
			},
			"dkim_host_name": rschema.StringAttribute{
				Description: "DKIM host name.",
				Optional:    true,
			},
			"force_root_dkim_host": rschema.BoolAttribute{
				Description: "Force using root domain for DKIM.",
				Optional:    true,
			},
			"encrypt_incoming_message": rschema.BoolAttribute{
				Description: "Whether to encrypt incoming messages.",
				Optional:    true,
			},

			// Computed attributes from API response
			"id": rschema.StringAttribute{
				Description: "Unique identifier of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": rschema.StringAttribute{
				Description: "Timestamp indicating when the domain was created.",
				Computed:    true,
			},
			"state": rschema.StringAttribute{
				Description: "Current verification status of the domain (active, unverified, disabled).",
				Computed:    true,
			},
			"smtp_login": rschema.StringAttribute{
				Description: "SMTP login username for the domain.",
				Computed:    true,
			},
			"is_disabled": rschema.BoolAttribute{
				Description: "Indicates whether the domain is currently disabled.",
				Computed:    true,
			},
			"require_tls": rschema.BoolAttribute{
				Description: "If true, Mailgun will only send messages over a TLS connection.",
				Computed:    true,
			},
			"skip_verification": rschema.BoolAttribute{
				Description: "If true, Mailgun will not verify the certificate and hostname when setting up a TLS connection.",
				Computed:    true,
			},
			"type": rschema.StringAttribute{
				Description: "Classification of the domain (custom or sandbox).",
				Computed:    true,
			},
			"tracking_host": rschema.StringAttribute{
				Description: "Custom tracking host for the domain used for tracking opens and clicks.",
				Computed:    true,
			},

			// DNS records
			"authentication_dns_records": rschema.ListNestedAttribute{
				Description: "Authentication DNS records generated by Mailgun (DMARC). Fetched from the Mailgun DMARC API (GET /v1/dmarc/records/{domain}), which is separate from the domain DNS records API.",
				Computed:    true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"cached": rschema.ListAttribute{
							Description: "Cached DNS values. Always empty for authentication records.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"is_active": rschema.BoolAttribute{
							Description: "Whether the DNS record is active. Always false for authentication records.",
							Computed:    true,
						},
						"name": rschema.StringAttribute{
							Description: "DNS record name (e.g. _dmarc.example.com).",
							Computed:    true,
						},
						"priority": rschema.StringAttribute{
							Description: "DNS record priority. Empty for authentication records.",
							Computed:    true,
						},
						"record_type": rschema.StringAttribute{
							Description: "DNS record type. Always TXT for authentication records.",
							Computed:    true,
						},
						"valid": rschema.StringAttribute{
							Description: "Whether the DNS record is valid. Derived from the Mailgun DMARC API `configured` field: \"valid\" when the DMARC record is detected in DNS, \"unknown\" otherwise.",
							Computed:    true,
						},
						"value": rschema.StringAttribute{
							Description: "DNS record value (the DMARC policy string).",
							Computed:    true,
						},
					},
				},
			},
			"receiving_dns_records": rschema.ListNestedAttribute{
				Description: "DNS records for receiving email.",
				Computed:    true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"cached": rschema.ListAttribute{
							Description: "Cached DNS values.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"is_active": rschema.BoolAttribute{
							Description: "Whether the DNS record is active.",
							Computed:    true,
						},
						"name": rschema.StringAttribute{
							Description: "DNS record name.",
							Computed:    true,
						},
						"priority": rschema.StringAttribute{
							Description: "DNS record priority.",
							Computed:    true,
						},
						"record_type": rschema.StringAttribute{
							Description: "DNS record type.",
							Computed:    true,
						},
						"valid": rschema.StringAttribute{
							Description: "Whether the DNS record is valid.",
							Computed:    true,
						},
						"value": rschema.StringAttribute{
							Description: "DNS record value.",
							Computed:    true,
						},
					},
				},
			},
			"sending_dns_records": rschema.ListNestedAttribute{
				Description: "DNS records for sending email.",
				Computed:    true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"cached": rschema.ListAttribute{
							Description: "Cached DNS values.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"is_active": rschema.BoolAttribute{
							Description: "Whether the DNS record is active.",
							Computed:    true,
						},
						"name": rschema.StringAttribute{
							Description: "DNS record name.",
							Computed:    true,
						},
						"priority": rschema.StringAttribute{
							Description: "DNS record priority.",
							Computed:    true,
						},
						"record_type": rschema.StringAttribute{
							Description: "DNS record type.",
							Computed:    true,
						},
						"valid": rschema.StringAttribute{
							Description: "Whether the DNS record is valid.",
							Computed:    true,
						},
						"value": rschema.StringAttribute{
							Description: "DNS record value.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// NewDomainResource creates a new domain resource
func NewDomainResource() resource.Resource {
	return &DomainResource{}
}
