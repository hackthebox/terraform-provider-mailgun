// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainDataSourceModel represents the Terraform state for a single domain data source
type DomainDataSourceModel struct {
	// Input - domain name to lookup
	Name types.String `tfsdk:"name"`

	// Computed attributes from the domain
	CreatedAt                  types.String `tfsdk:"created_at"`
	Id                         types.String `tfsdk:"id"`
	IsDisabled                 types.Bool   `tfsdk:"is_disabled"`
	RequireTls                 types.Bool   `tfsdk:"require_tls"`
	SkipVerification           types.Bool   `tfsdk:"skip_verification"`
	SmtpLogin                  types.String `tfsdk:"smtp_login"`
	SmtpPassword               types.String `tfsdk:"smtp_password"`
	SpamAction                 types.String `tfsdk:"spam_action"`
	State                      types.String `tfsdk:"state"`
	TrackingHost               types.String `tfsdk:"tracking_host"`
	DomainType                 types.String `tfsdk:"type"`
	UseAutomaticSenderSecurity types.Bool   `tfsdk:"use_automatic_sender_security"`
	WebPrefix                  types.String `tfsdk:"web_prefix"`
	WebScheme                  types.String `tfsdk:"web_scheme"`
	Wildcard                   types.Bool   `tfsdk:"wildcard"`
	ReceivingDnsRecords        types.List   `tfsdk:"receiving_dns_records"`
	SendingDnsRecords          types.List   `tfsdk:"sending_dns_records"`
	AuthenticationDnsRecords   types.List   `tfsdk:"authentication_dns_records"`
}
