// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainDkimKeysDataSourceModel represents the data source state for listing Mailgun domain DKIM keys.
type DomainDkimKeysDataSourceModel struct {
	Domain types.String       `tfsdk:"domain"`
	Keys   []DkimKeyItemModel `tfsdk:"keys"`
}

// DkimKeyItemModel represents a single DKIM key in the list.
type DkimKeyItemModel struct {
	Selector       types.String `tfsdk:"selector"`
	SigningDomain  types.String `tfsdk:"signing_domain"`
	Active         types.Bool   `tfsdk:"active"`
	DnsRecordName  types.String `tfsdk:"dns_record_name"`
	DnsRecordType  types.String `tfsdk:"dns_record_type"`
	DnsRecordValue types.String `tfsdk:"dns_record_value"`
	DnsRecordValid types.String `tfsdk:"dns_record_valid"`
}
