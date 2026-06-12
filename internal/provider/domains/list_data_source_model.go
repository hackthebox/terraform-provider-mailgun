// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// DomainsModel represents the Terraform state for the domains list data source
type DomainsModel struct {
	Authority          types.String `tfsdk:"authority"`
	IncludeSubaccounts types.Bool   `tfsdk:"include_subaccounts"`
	Limit              types.Int64  `tfsdk:"limit"`
	Search             types.String `tfsdk:"search"`
	Skip               types.Int64  `tfsdk:"skip"`
	Sort               types.String `tfsdk:"sort"`
	State              types.String `tfsdk:"state"`
	TotalCount         types.Int64  `tfsdk:"total_count"`
	Items              types.List   `tfsdk:"items"`
}

// ItemsValue represents a domain item in the list data source
type ItemsValue struct {
	CreatedAt                  types.String `tfsdk:"created_at"`
	Disabled                   types.Object `tfsdk:"disabled"`
	Id                         types.String `tfsdk:"id"`
	IsDisabled                 types.Bool   `tfsdk:"is_disabled"`
	Name                       types.String `tfsdk:"name"`
	RequireTls                 types.Bool   `tfsdk:"require_tls"`
	SkipVerification           types.Bool   `tfsdk:"skip_verification"`
	SmtpLogin                  types.String `tfsdk:"smtp_login"`
	SmtpPassword               types.String `tfsdk:"smtp_password"`
	SpamAction                 types.String `tfsdk:"spam_action"`
	State                      types.String `tfsdk:"state"`
	TrackingHost               types.String `tfsdk:"tracking_host"`
	ItemsType                  types.String `tfsdk:"type"`
	UseAutomaticSenderSecurity types.Bool   `tfsdk:"use_automatic_sender_security"`
	WebPrefix                  types.String `tfsdk:"web_prefix"`
	WebScheme                  types.String `tfsdk:"web_scheme"`
	Wildcard                   types.Bool   `tfsdk:"wildcard"`
	state                      attr.ValueState
}

// AttributeTypes returns the attribute types for ItemsValue
func (v ItemsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"created_at": types.StringType,
		"disabled": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"code":        types.StringType,
				"note":        types.StringType,
				"permanently": types.BoolType,
				"reason":      types.StringType,
				"until":       types.StringType,
			},
		},
		"id":                            types.StringType,
		"is_disabled":                   types.BoolType,
		"name":                          types.StringType,
		"require_tls":                   types.BoolType,
		"skip_verification":             types.BoolType,
		"smtp_login":                    types.StringType,
		"smtp_password":                 types.StringType,
		"spam_action":                   types.StringType,
		"state":                         types.StringType,
		"tracking_host":                 types.StringType,
		"type":                          types.StringType,
		"use_automatic_sender_security": types.BoolType,
		"web_prefix":                    types.StringType,
		"web_scheme":                    types.StringType,
		"wildcard":                      types.BoolType,
	}
}

// ToObjectValue converts ItemsValue to an ObjectValue
func (v ItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(v.AttributeTypes(ctx), map[string]attr.Value{
		"created_at":                    v.CreatedAt,
		"disabled":                      v.Disabled,
		"id":                            v.Id,
		"is_disabled":                   v.IsDisabled,
		"name":                          v.Name,
		"require_tls":                   v.RequireTls,
		"skip_verification":             v.SkipVerification,
		"smtp_login":                    v.SmtpLogin,
		"smtp_password":                 v.SmtpPassword,
		"spam_action":                   v.SpamAction,
		"state":                         v.State,
		"tracking_host":                 v.TrackingHost,
		"type":                          v.ItemsType,
		"use_automatic_sender_security": v.UseAutomaticSenderSecurity,
		"web_prefix":                    v.WebPrefix,
		"web_scheme":                    v.WebScheme,
		"wildcard":                      v.Wildcard,
	})
}
