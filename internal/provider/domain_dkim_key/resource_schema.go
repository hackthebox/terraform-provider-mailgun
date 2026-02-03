// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// DomainDkimKeyResourceSchema returns the schema for the mailgun_domain_dkim_key resource.
func DomainDkimKeyResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages DKIM keys for a Mailgun domain. DKIM (DomainKeys Identified Mail) allows email recipients to verify that messages were sent from an authorized mail server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this DKIM key (domain/selector).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The domain name for this DKIM key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"selector": schema.StringAttribute{
				Description: "The DKIM selector. This is used in the DNS record name (e.g., selector._domainkey.domain.com).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bits": schema.Int64Attribute{
				Description: "The key size in bits. Valid values are 1024 or 2048. Defaults to 1024.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1024),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the DKIM key is active. Defaults to false. Note: Activation requires valid DNS records and is not supported on sandbox domains.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"signing_domain": schema.StringAttribute{
				Description: "The domain used for signing emails.",
				Computed:    true,
			},
			"dns_record_name": schema.StringAttribute{
				Description: "The DNS record name to configure for DKIM.",
				Computed:    true,
			},
			"dns_record_type": schema.StringAttribute{
				Description: "The DNS record type (typically TXT).",
				Computed:    true,
			},
			"dns_record_value": schema.StringAttribute{
				Description: "The DNS record value containing the DKIM public key.",
				Computed:    true,
			},
			"dns_record_valid": schema.StringAttribute{
				Description: "Whether the DNS record has been validated (valid/invalid/unknown).",
				Computed:    true,
			},
		},
	}
}
