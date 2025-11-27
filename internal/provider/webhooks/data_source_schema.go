// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WebhooksDataSourceSchema returns the schema for the mailgun_webhooks data source.
func WebhooksDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Lists all Mailgun webhooks for a domain. Webhooks allow you to receive HTTP POST requests for email events.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to list webhooks for.",
				Required:    true,
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of webhooks returned.",
				Computed:    true,
			},
			"webhooks": schema.ListNestedAttribute{
				Description: "List of webhooks configured for the domain.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"webhook_type": schema.StringAttribute{
							Description: "The type of webhook event (accepted, delivered, permanent_fail, temporary_fail, opened, clicked, unsubscribed, complained).",
							Computed:    true,
						},
						"urls": schema.ListAttribute{
							Description: "List of URLs that receive webhook POST requests for this event type.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}
