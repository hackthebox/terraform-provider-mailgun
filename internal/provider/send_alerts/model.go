// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SendAlertModel represents a Mailgun send alert in Terraform state.
type SendAlertModel struct {
	// Computed fields
	ID              types.String `tfsdk:"id"`
	ParentAccountID types.String `tfsdk:"parent_account_id"`
	SubaccountID    types.String `tfsdk:"subaccount_id"`
	AccountGroup    types.String `tfsdk:"account_group"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	LastChecked     types.String `tfsdk:"last_checked"`

	// Required fields
	Name       types.String `tfsdk:"name"`
	Metric     types.String `tfsdk:"metric"`
	Comparator types.String `tfsdk:"comparator"`
	Limit      types.String `tfsdk:"limit"`
	Dimension  types.String `tfsdk:"dimension"`

	// Optional fields
	Description   types.String `tfsdk:"description"`
	Period        types.String `tfsdk:"period"`
	AlertChannels types.List   `tfsdk:"alert_channels"`
	Filters       types.List   `tfsdk:"filters"`
}

// SendAlertFilterModel represents a filter in a send alert.
type SendAlertFilterModel struct {
	Dimension  types.String `tfsdk:"dimension"`
	Comparator types.String `tfsdk:"comparator"`
	Values     types.List   `tfsdk:"values"`
}

// SendAlertsListModel represents the state for the send alerts list data source.
type SendAlertsListModel struct {
	Alerts     types.List  `tfsdk:"alerts"`
	TotalCount types.Int64 `tfsdk:"total_count"`
}

// API Request/Response types

// SendAlertAPIRequest represents the request body for creating/updating a send alert.
type SendAlertAPIRequest struct {
	Name          string                     `json:"name"`
	Metric        string                     `json:"metric"`
	Comparator    string                     `json:"comparator"`
	Limit         string                     `json:"limit"`
	Dimension     string                     `json:"dimension"`
	Description   string                     `json:"description,omitempty"`
	Period        string                     `json:"period,omitempty"`
	AlertChannels []string                   `json:"alert_channels,omitempty"`
	Filters       []SendAlertFilterAPIObject `json:"filters,omitempty"`
}

// SendAlertFilterAPIObject represents a filter in the API request/response.
type SendAlertFilterAPIObject struct {
	Dimension  string   `json:"dimension"`
	Comparator string   `json:"comparator,omitempty"`
	Values     []string `json:"values"`
}

// SendAlertAPIResponse represents the response from the API when getting a single alert.
type SendAlertAPIResponse struct {
	ID              string                     `json:"id,omitempty"`
	ParentAccountID string                     `json:"parent_account_id,omitempty"`
	SubaccountID    string                     `json:"subaccount_id,omitempty"`
	AccountGroup    string                     `json:"account_group,omitempty"`
	Name            string                     `json:"name"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at,omitempty"`
	LastChecked     string                     `json:"last_checked,omitempty"`
	Description     string                     `json:"description,omitempty"`
	AlertChannels   []string                   `json:"alert_channels,omitempty"`
	Filters         []SendAlertFilterAPIObject `json:"filters,omitempty"`
	Metric          string                     `json:"metric"`
	Limit           string                     `json:"limit"`
	Dimension       string                     `json:"dimension"`
	Period          string                     `json:"period,omitempty"`
	Comparator      string                     `json:"comparator,omitempty"`
}

// SendAlertsListAPIResponse represents the response from listing send alerts.
type SendAlertsListAPIResponse struct {
	Items []SendAlertAPIResponse `json:"items"`
	Total int                    `json:"total"`
}

// MessageResponse represents a simple message response from the API.
type MessageResponse struct {
	Message string `json:"message"`
}
