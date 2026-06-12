// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// DomainsDataSourceSchema returns the schema for the domains list data source
func DomainsListDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves a list of all domains in your Mailgun account.",
		Attributes: map[string]schema.Attribute{
			"authority": schema.StringAttribute{
				Description: "Filter by authority.",
				Optional:    true,
			},
			"include_subaccounts": schema.BoolAttribute{
				Description: "Whether to include subaccounts.",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of domains to return (default: 100).",
				Optional:    true,
			},
			"search": schema.StringAttribute{
				Description: "Search query to filter domains.",
				Optional:    true,
			},
			"skip": schema.Int64Attribute{
				Description: "Number of domains to skip.",
				Optional:    true,
			},
			"sort": schema.StringAttribute{
				Description: "Sort order for results.",
				Optional:    true,
			},
			"state": schema.StringAttribute{
				Description: "Filter by domain state.",
				Optional:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of domains returned.",
				Computed:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of domains.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"created_at": schema.StringAttribute{
							Description: "Time the domain was created.",
							Computed:    true,
						},
						"disabled": schema.SingleNestedAttribute{
							Description: "Information about domain being disabled.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"code": schema.StringAttribute{
									Description: "Disable code.",
									Computed:    true,
								},
								"note": schema.StringAttribute{
									Description: "Disable note.",
									Computed:    true,
								},
								"permanently": schema.BoolAttribute{
									Description: "Whether disabled permanently.",
									Computed:    true,
								},
								"reason": schema.StringAttribute{
									Description: "Disable reason.",
									Computed:    true,
								},
								"until": schema.StringAttribute{
									Description: "Disabled until timestamp.",
									Computed:    true,
								},
							},
						},
						"id": schema.StringAttribute{
							Description: "Domain ID.",
							Computed:    true,
						},
						"is_disabled": schema.BoolAttribute{
							Description: "Whether the domain is disabled.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Domain name.",
							Computed:    true,
						},
						"require_tls": schema.BoolAttribute{
							Description: "Whether TLS is required.",
							Computed:    true,
						},
						"skip_verification": schema.BoolAttribute{
							Description: "Whether to skip verification.",
							Computed:    true,
						},
						"smtp_login": schema.StringAttribute{
							Description: "SMTP login username.",
							Computed:    true,
						},
						"smtp_password": schema.StringAttribute{
							Description: "SMTP password.",
							Computed:    true,
							Sensitive:   true,
						},
						"spam_action": schema.StringAttribute{
							Description: "Spam action setting.",
							Computed:    true,
						},
						"state": schema.StringAttribute{
							Description: "Domain state.",
							Computed:    true,
						},
						"tracking_host": schema.StringAttribute{
							Description: "Tracking host.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Domain type.",
							Computed:    true,
						},
						"use_automatic_sender_security": schema.BoolAttribute{
							Description: "Whether automatic sender security is enabled.",
							Computed:    true,
						},
						"web_prefix": schema.StringAttribute{
							Description: "Web prefix for tracking.",
							Computed:    true,
						},
						"web_scheme": schema.StringAttribute{
							Description: "Web scheme (http or https).",
							Computed:    true,
						},
						"wildcard": schema.BoolAttribute{
							Description: "Whether wildcard is enabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// NewDomainsListDataSource creates a new domains list data source
func NewDomainsListDataSource() datasource.DataSource {
	return &ListDataSource{}
}
