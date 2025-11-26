// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// ApiKeyResourceSchema returns the schema for the API key resource
func ApiKeyResourceSchema() rschema.Schema {
	return rschema.Schema{
		Description: "Manages a Mailgun API key. API keys provide programmatic access to Mailgun services.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Description: "The unique identifier for this API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": rschema.StringAttribute{
				Description: "The role for this API key. Valid values: 'admin', 'sending', 'developer', 'basic', 'support'. " +
					"For domain-specific sending keys, use 'sending' with 'kind' set to 'domain'.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": rschema.StringAttribute{
				Description: "A human-readable description for this API key.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_name": rschema.StringAttribute{
				Description: "The domain this API key is scoped to. Required when 'role' is 'sending' and 'kind' is 'domain'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": rschema.StringAttribute{
				Description: "The type of API key. Valid values: 'domain' (for domain-specific sending keys), " +
					"'user' (default, for user-level keys), 'web' (for web keys with 1-day max lifetime).",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("user"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration": rschema.Int64Attribute{
				Description: "The key's lifetime in seconds. Set to 0 for no expiration. Web keys have a maximum lifetime of 1 day.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"secret": rschema.StringAttribute{
				Description: "The API key secret. This is only available immediately after creation and cannot be retrieved later. " +
					"Store this value securely as it will not be shown again.",
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": rschema.StringAttribute{
				Description: "The timestamp when this API key was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": rschema.StringAttribute{
				Description: "The timestamp when this API key was last updated.",
				Computed:    true,
			},
			"expires_at": rschema.StringAttribute{
				Description: "The timestamp when this API key expires. Empty if the key does not expire.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_disabled": rschema.BoolAttribute{
				Description: "Whether this API key is disabled.",
				Computed:    true,
			},
			"disabled_reason": rschema.StringAttribute{
				Description: "The reason this API key was disabled, if applicable.",
				Computed:    true,
			},
			"requestor": rschema.StringAttribute{
				Description: "The entity that requested this API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_name": rschema.StringAttribute{
				Description: "The username associated with this API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// NewApiKeyResource creates a new API key resource
func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}
