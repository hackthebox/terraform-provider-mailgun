// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package alerts

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlertDataSourceSchema returns the schema for the mailgun_alert data source.
func AlertDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Use this data source to look up an existing Mailgun alert by ID. " +
			"Alerts notify you when specific events occur in your Mailgun account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the alert (UUID format).",
				Required:    true,
			},
			"event_type": schema.StringAttribute{
				Description: "The type of event that triggers the alert.",
				Computed:    true,
			},
			"channel": schema.StringAttribute{
				Description: "The delivery channel for the alert: 'email', 'webhook', or 'slack'.",
				Computed:    true,
			},
			"emails": schema.ListAttribute{
				Description: "List of email addresses receiving alert notifications (when channel is 'email').",
				Computed:    true,
				ElementType: types.StringType,
			},
			"webhook_url": schema.StringAttribute{
				Description: "The URL receiving webhook notifications (when channel is 'webhook').",
				Computed:    true,
			},
			"slack_ids": schema.ListAttribute{
				Description: "List of Slack channel IDs receiving alert notifications (when channel is 'slack').",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// AlertsListDataSourceSchema returns the schema for the mailgun_alerts data source.
func AlertsListDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Use this data source to list all Mailgun alerts configured for your account. " +
			"Alerts notify you when specific events occur in your Mailgun account.",
		Attributes: map[string]schema.Attribute{
			"alerts": schema.ListNestedAttribute{
				Description: "List of configured alerts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the alert (UUID format).",
							Computed:    true,
						},
						"event_type": schema.StringAttribute{
							Description: "The type of event that triggers the alert.",
							Computed:    true,
						},
						"channel": schema.StringAttribute{
							Description: "The delivery channel for the alert: 'email', 'webhook', or 'slack'.",
							Computed:    true,
						},
						"emails": schema.ListAttribute{
							Description: "List of email addresses receiving alert notifications (when channel is 'email').",
							Computed:    true,
							ElementType: types.StringType,
						},
						"webhook_url": schema.StringAttribute{
							Description: "The URL receiving webhook notifications (when channel is 'webhook').",
							Computed:    true,
						},
						"slack_ids": schema.ListAttribute{
							Description: "List of Slack channel IDs receiving alert notifications (when channel is 'slack').",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"total_count": schema.Int64Attribute{
				Description: "The total number of alerts returned.",
				Computed:    true,
			},
		},
	}
}

// AlertEventsDataSourceSchema returns the schema for the mailgun_alert_events data source.
func AlertEventsDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Use this data source to list all available alert event types that can be used " +
			"when creating alerts. Event types define what triggers an alert notification.",
		Attributes: map[string]schema.Attribute{
			"events": schema.ListAttribute{
				Description: "List of available event types for alerts.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}
