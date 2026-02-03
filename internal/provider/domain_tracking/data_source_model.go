// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainTrackingDataSourceModel represents the data source state for Mailgun domain tracking settings.
type DomainTrackingDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`

	// Click tracking
	ClickActive types.Bool `tfsdk:"click_active"`

	// Open tracking
	OpenActive types.Bool `tfsdk:"open_active"`

	// Unsubscribe tracking
	UnsubscribeActive     types.Bool   `tfsdk:"unsubscribe_active"`
	UnsubscribeHtmlFooter types.String `tfsdk:"unsubscribe_html_footer"`
	UnsubscribeTextFooter types.String `tfsdk:"unsubscribe_text_footer"`
}
