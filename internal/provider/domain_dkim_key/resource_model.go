// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainDkimKeyModel represents the resource state for a Mailgun domain DKIM key.
type DomainDkimKeyModel struct {
	Id            types.String `tfsdk:"id"`
	Domain        types.String `tfsdk:"domain"`
	Selector      types.String `tfsdk:"selector"`
	Bits          types.Int64  `tfsdk:"bits"`
	Active        types.Bool   `tfsdk:"active"`
	SigningDomain types.String `tfsdk:"signing_domain"`

	// DNS Record details (computed)
	DnsRecordName  types.String `tfsdk:"dns_record_name"`
	DnsRecordType  types.String `tfsdk:"dns_record_type"`
	DnsRecordValue types.String `tfsdk:"dns_record_value"`
	DnsRecordValid types.String `tfsdk:"dns_record_valid"`
}
