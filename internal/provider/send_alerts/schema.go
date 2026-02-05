// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Valid metric values for send alerts.
var ValidMetrics = []string{
	"hard_bounce_rate",
	"temporary_fail_rate",
	"delivered_rate",
	"complained_rate",
}

// Valid comparator values for send alerts.
var ValidComparators = []string{
	"=",
	"!=",
	"<",
	"<=",
	">",
	">=",
}

// Valid dimension values for send alerts.
var ValidDimensions = []string{
	"domain",
	"ip",
	"ip_pool",
	"recipient_provider",
	"subaccount",
}

// Valid alert channel values.
var ValidAlertChannels = []string{
	"email",
	"slack",
	"webhook",
}

// FilterObjectType returns the object type for filter attributes.
func FilterObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"dimension":  types.StringType,
		"comparator": types.StringType,
		"values":     types.ListType{ElemType: types.StringType},
	}
}

// SendAlertResourceSchema returns the schema for the send_alert resource.
func SendAlertResourceSchema() resourceschema.Schema {
	return resourceschema.Schema{
		Description: "Manages a Mailgun send alert. Send alerts notify you when sending metrics cross defined thresholds.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				Description: "The unique identifier for the alert.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_account_id": resourceschema.StringAttribute{
				Description: "The parent account ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subaccount_id": resourceschema.StringAttribute{
				Description: "The subaccount ID this alert belongs to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_group": resourceschema.StringAttribute{
				Description: "The group this account belongs to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": resourceschema.StringAttribute{
				Description: "Timestamp of when the alert was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": resourceschema.StringAttribute{
				Description: "Timestamp of when the alert was last updated.",
				Computed:    true,
			},
			"last_checked": resourceschema.StringAttribute{
				Description: "Timestamp of when the alert was last checked.",
				Computed:    true,
			},
			"name": resourceschema.StringAttribute{
				Description: "A user-friendly name for the alert. This is used as the unique identifier.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metric": resourceschema.StringAttribute{
				Description: "The metric being monitored. Valid values: hard_bounce_rate, temporary_fail_rate, delivered_rate, complained_rate.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(ValidMetrics...),
				},
			},
			"comparator": resourceschema.StringAttribute{
				Description: "The comparison operator. Valid values: =, !=, <, <=, >, >=.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(ValidComparators...),
				},
			},
			"limit": resourceschema.StringAttribute{
				Description: "The threshold limit for the alert (e.g., '0.05' for 5%).",
				Required:    true,
			},
			"dimension": resourceschema.StringAttribute{
				Description: "The dimension to apply to the metric. Valid values: domain, ip, ip_pool, recipient_provider, subaccount.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(ValidDimensions...),
				},
			},
			"description": resourceschema.StringAttribute{
				Description: "A description of what the alert does.",
				Optional:    true,
			},
			"period": resourceschema.StringAttribute{
				Description: "The time period for the metric aggregation (e.g., '1h', '1d').",
				Optional:    true,
			},
			"alert_channels": resourceschema.ListAttribute{
				Description: "A list of alert channels to notify. Valid values: email, slack, webhook.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"filters": resourceschema.ListNestedAttribute{
				Description: "A list of filters to apply to the alert.",
				Optional:    true,
				NestedObject: resourceschema.NestedAttributeObject{
					Attributes: map[string]resourceschema.Attribute{
						"dimension": resourceschema.StringAttribute{
							Description: "The dimension to filter by.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(ValidDimensions...),
							},
						},
						"comparator": resourceschema.StringAttribute{
							Description: "The comparison operator for the filter.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(ValidComparators...),
							},
						},
						"values": resourceschema.ListAttribute{
							Description: "The dimension values to apply to filter.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// SendAlertDataSourceSchema returns the schema for the send_alert data source.
func SendAlertDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Fetches a Mailgun send alert by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the alert.",
				Computed:    true,
			},
			"parent_account_id": schema.StringAttribute{
				Description: "The parent account ID.",
				Computed:    true,
			},
			"subaccount_id": schema.StringAttribute{
				Description: "The subaccount ID this alert belongs to.",
				Computed:    true,
			},
			"account_group": schema.StringAttribute{
				Description: "The group this account belongs to.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the alert was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp of when the alert was last updated.",
				Computed:    true,
			},
			"last_checked": schema.StringAttribute{
				Description: "Timestamp of when the alert was last checked.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the alert to fetch.",
				Required:    true,
			},
			"metric": schema.StringAttribute{
				Description: "The metric being monitored.",
				Computed:    true,
			},
			"comparator": schema.StringAttribute{
				Description: "The comparison operator.",
				Computed:    true,
			},
			"limit": schema.StringAttribute{
				Description: "The threshold limit for the alert.",
				Computed:    true,
			},
			"dimension": schema.StringAttribute{
				Description: "The dimension to apply to the metric.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of what the alert does.",
				Computed:    true,
			},
			"period": schema.StringAttribute{
				Description: "The time period for the metric aggregation.",
				Computed:    true,
			},
			"alert_channels": schema.ListAttribute{
				Description: "A list of alert channels to notify.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"filters": schema.ListNestedAttribute{
				Description: "A list of filters applied to the alert.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dimension": schema.StringAttribute{
							Description: "The dimension to filter by.",
							Computed:    true,
						},
						"comparator": schema.StringAttribute{
							Description: "The comparison operator for the filter.",
							Computed:    true,
						},
						"values": schema.ListAttribute{
							Description: "The dimension values to apply to filter.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// SendAlertsListDataSourceSchema returns the schema for the send_alerts list data source.
func SendAlertsListDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Lists all Mailgun send alerts for the account.",
		Attributes: map[string]schema.Attribute{
			"total_count": schema.Int64Attribute{
				Description: "Total number of send alerts.",
				Computed:    true,
			},
			"alerts": schema.ListNestedAttribute{
				Description: "List of send alerts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the alert.",
							Computed:    true,
						},
						"parent_account_id": schema.StringAttribute{
							Description: "The parent account ID.",
							Computed:    true,
						},
						"subaccount_id": schema.StringAttribute{
							Description: "The subaccount ID this alert belongs to.",
							Computed:    true,
						},
						"account_group": schema.StringAttribute{
							Description: "The group this account belongs to.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp of when the alert was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp of when the alert was last updated.",
							Computed:    true,
						},
						"last_checked": schema.StringAttribute{
							Description: "Timestamp of when the alert was last checked.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "A user-friendly name for the alert.",
							Computed:    true,
						},
						"metric": schema.StringAttribute{
							Description: "The metric being monitored.",
							Computed:    true,
						},
						"comparator": schema.StringAttribute{
							Description: "The comparison operator.",
							Computed:    true,
						},
						"limit": schema.StringAttribute{
							Description: "The threshold limit for the alert.",
							Computed:    true,
						},
						"dimension": schema.StringAttribute{
							Description: "The dimension to apply to the metric.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of what the alert does.",
							Computed:    true,
						},
						"period": schema.StringAttribute{
							Description: "The time period for the metric aggregation.",
							Computed:    true,
						},
						"alert_channels": schema.ListAttribute{
							Description: "A list of alert channels to notify.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"filters": schema.ListNestedAttribute{
							Description: "A list of filters applied to the alert.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"dimension": schema.StringAttribute{
										Description: "The dimension to filter by.",
										Computed:    true,
									},
									"comparator": schema.StringAttribute{
										Description: "The comparison operator for the filter.",
										Computed:    true,
									},
									"values": schema.ListAttribute{
										Description: "The dimension values to apply to filter.",
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
