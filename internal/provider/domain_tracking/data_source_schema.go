// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// DomainTrackingDataSourceSchema returns the schema for the mailgun_domain_tracking data source.
func DomainTrackingDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves tracking settings for a Mailgun domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to retrieve tracking settings for.",
				Required:    true,
			},
			"click_active": schema.BoolAttribute{
				Description: "Whether click tracking is enabled.",
				Computed:    true,
			},
			"open_active": schema.BoolAttribute{
				Description: "Whether open tracking is enabled.",
				Computed:    true,
			},
			"unsubscribe_active": schema.BoolAttribute{
				Description: "Whether unsubscribe tracking is enabled.",
				Computed:    true,
			},
			"unsubscribe_html_footer": schema.StringAttribute{
				Description: "Custom HTML footer for unsubscribe links.",
				Computed:    true,
			},
			"unsubscribe_text_footer": schema.StringAttribute{
				Description: "Custom text footer for unsubscribe links.",
				Computed:    true,
			},
		},
	}
}
