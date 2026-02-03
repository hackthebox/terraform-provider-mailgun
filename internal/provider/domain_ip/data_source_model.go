// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainIPsDataSourceModel represents the data source state for listing Mailgun domain IPs.
type DomainIPsDataSourceModel struct {
	Domain types.String  `tfsdk:"domain"`
	IPs    []IPItemModel `tfsdk:"ips"`
}

// IPItemModel represents a single IP address in the list.
type IPItemModel struct {
	IP types.String `tfsdk:"ip"`
}
