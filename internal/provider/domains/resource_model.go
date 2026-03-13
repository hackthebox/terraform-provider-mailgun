// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainModel represents the Terraform state for a domain resource
type DomainModel struct {
	// Required/Optional input attributes
	Name                       types.String `tfsdk:"name"`
	SmtpPassword               types.String `tfsdk:"smtp_password"`
	SpamAction                 types.String `tfsdk:"spam_action"`
	Wildcard                   types.Bool   `tfsdk:"wildcard"`
	ForceDkimAuthority         types.Bool   `tfsdk:"force_dkim_authority"`
	DkimKeySize                types.String `tfsdk:"dkim_key_size"`
	Ips                        types.String `tfsdk:"ips"`
	PoolId                     types.String `tfsdk:"pool_id"`
	WebScheme                  types.String `tfsdk:"web_scheme"`
	WebPrefix                  types.String `tfsdk:"web_prefix"`
	UseAutomaticSenderSecurity types.Bool   `tfsdk:"use_automatic_sender_security"`
	DkimSelector               types.String `tfsdk:"dkim_selector"`
	DkimHostName               types.String `tfsdk:"dkim_host_name"`
	ForceRootDkimHost          types.Bool   `tfsdk:"force_root_dkim_host"`
	EncryptIncomingMessage     types.Bool   `tfsdk:"encrypt_incoming_message"`

	// Computed attributes from API response (flat, not nested)
	Id               types.String `tfsdk:"id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	State            types.String `tfsdk:"state"`
	SmtpLogin        types.String `tfsdk:"smtp_login"`
	IsDisabled       types.Bool   `tfsdk:"is_disabled"`
	RequireTls       types.Bool   `tfsdk:"require_tls"`
	SkipVerification types.Bool   `tfsdk:"skip_verification"`
	DomainType       types.String `tfsdk:"type"`
	TrackingHost     types.String `tfsdk:"tracking_host"`

	// DNS records
	ReceivingDnsRecords      types.List `tfsdk:"receiving_dns_records"`
	SendingDnsRecords        types.List `tfsdk:"sending_dns_records"`
	AuthenticationDnsRecords types.List `tfsdk:"authentication_dns_records"`
}

// ReceivingDnsRecordsValue represents a DNS record for receiving
type ReceivingDnsRecordsValue struct {
	Cached     types.List   `tfsdk:"cached"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	Name       types.String `tfsdk:"name"`
	Priority   types.String `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
	state      attr.ValueState
}

// AttributeTypes returns the attribute types for ReceivingDnsRecordsValue
func (v ReceivingDnsRecordsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"cached":      types.ListType{ElemType: types.StringType},
		"is_active":   types.BoolType,
		"name":        types.StringType,
		"priority":    types.StringType,
		"record_type": types.StringType,
		"valid":       types.StringType,
		"value":       types.StringType,
	}
}

// ReceivingDnsRecordsType represents the list type for receiving DNS records
type ReceivingDnsRecordsType struct {
	types.ObjectType
}

// SendingDnsRecordsValue represents a DNS record for sending
type SendingDnsRecordsValue struct {
	Cached     types.List   `tfsdk:"cached"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	Name       types.String `tfsdk:"name"`
	Priority   types.String `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
	state      attr.ValueState
}

// AttributeTypes returns the attribute types for SendingDnsRecordsValue
func (v SendingDnsRecordsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"cached":      types.ListType{ElemType: types.StringType},
		"is_active":   types.BoolType,
		"name":        types.StringType,
		"priority":    types.StringType,
		"record_type": types.StringType,
		"valid":       types.StringType,
		"value":       types.StringType,
	}
}

// SendingDnsRecordsType represents the list type for sending DNS records
type SendingDnsRecordsType struct {
	types.ObjectType
}

// AuthenticationDnsRecordsValue represents a DNS record for authentication (e.g. DMARC)
type AuthenticationDnsRecordsValue struct {
	Cached     types.List   `tfsdk:"cached"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	Name       types.String `tfsdk:"name"`
	Priority   types.String `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
	state      attr.ValueState
}

// AttributeTypes returns the attribute types for AuthenticationDnsRecordsValue
func (v AuthenticationDnsRecordsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"cached":      types.ListType{ElemType: types.StringType},
		"is_active":   types.BoolType,
		"name":        types.StringType,
		"priority":    types.StringType,
		"record_type": types.StringType,
		"valid":       types.StringType,
		"value":       types.StringType,
	}
}

// AuthenticationDnsRecordsType represents the list type for authentication DNS records
type AuthenticationDnsRecordsType struct {
	types.ObjectType
}
