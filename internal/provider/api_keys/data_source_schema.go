// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// ApiKeyDataSourceSchema returns the schema for the single API key data source
func ApiKeyDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		Description: "Retrieves information about a specific Mailgun API key.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Description: "The unique identifier for the API key.",
				Required:    true,
			},
			"role": dschema.StringAttribute{
				Description: "The role of this API key.",
				Computed:    true,
			},
			"kind": dschema.StringAttribute{
				Description: "The type of API key (domain, user, or web).",
				Computed:    true,
			},
			"description": dschema.StringAttribute{
				Description: "The description of this API key.",
				Computed:    true,
			},
			"domain_name": dschema.StringAttribute{
				Description: "The domain this API key is scoped to (if applicable).",
				Computed:    true,
			},
			"created_at": dschema.StringAttribute{
				Description: "The timestamp when this API key was created.",
				Computed:    true,
			},
			"updated_at": dschema.StringAttribute{
				Description: "The timestamp when this API key was last updated.",
				Computed:    true,
			},
			"expires_at": dschema.StringAttribute{
				Description: "The timestamp when this API key expires.",
				Computed:    true,
			},
			"is_disabled": dschema.BoolAttribute{
				Description: "Whether this API key is disabled.",
				Computed:    true,
			},
			"disabled_reason": dschema.StringAttribute{
				Description: "The reason this API key was disabled.",
				Computed:    true,
			},
			"requestor": dschema.StringAttribute{
				Description: "The entity that requested this API key.",
				Computed:    true,
			},
			"user_name": dschema.StringAttribute{
				Description: "The username associated with this API key.",
				Computed:    true,
			},
		},
	}
}

// ApiKeysListDataSourceSchema returns the schema for the API keys list data source
func ApiKeysListDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		Description: "Lists Mailgun API keys with optional filtering.",
		Attributes: map[string]dschema.Attribute{
			"domain_name": dschema.StringAttribute{
				Description: "Filter API keys by domain name.",
				Optional:    true,
			},
			"kind": dschema.StringAttribute{
				Description: "Filter API keys by kind (domain, user, or web).",
				Optional:    true,
			},
			"total_count": dschema.Int64Attribute{
				Description: "The total number of API keys returned.",
				Computed:    true,
			},
			"keys": dschema.ListNestedAttribute{
				Description: "List of API keys.",
				Computed:    true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"id": dschema.StringAttribute{
							Description: "The unique identifier for this API key.",
							Computed:    true,
						},
						"role": dschema.StringAttribute{
							Description: "The role of this API key.",
							Computed:    true,
						},
						"kind": dschema.StringAttribute{
							Description: "The type of API key.",
							Computed:    true,
						},
						"description": dschema.StringAttribute{
							Description: "The description of this API key.",
							Computed:    true,
						},
						"domain_name": dschema.StringAttribute{
							Description: "The domain this API key is scoped to.",
							Computed:    true,
						},
						"created_at": dschema.StringAttribute{
							Description: "The timestamp when this API key was created.",
							Computed:    true,
						},
						"updated_at": dschema.StringAttribute{
							Description: "The timestamp when this API key was last updated.",
							Computed:    true,
						},
						"expires_at": dschema.StringAttribute{
							Description: "The timestamp when this API key expires.",
							Computed:    true,
						},
						"is_disabled": dschema.BoolAttribute{
							Description: "Whether this API key is disabled.",
							Computed:    true,
						},
						"disabled_reason": dschema.StringAttribute{
							Description: "The reason this API key was disabled.",
							Computed:    true,
						},
						"requestor": dschema.StringAttribute{
							Description: "The entity that requested this API key.",
							Computed:    true,
						},
						"user_name": dschema.StringAttribute{
							Description: "The username associated with this API key.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// NewApiKeyDataSource creates a new single API key data source
func NewApiKeyDataSource() datasource.DataSource {
	return &ApiKeyDataSource{}
}

// NewApiKeysListDataSource creates a new API keys list data source
func NewApiKeysListDataSource() datasource.DataSource {
	return &ApiKeysListDataSource{}
}
