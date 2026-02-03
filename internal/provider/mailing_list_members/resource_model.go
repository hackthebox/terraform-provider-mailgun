// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MailingListMemberModel represents the Terraform state for a Mailgun mailing list member.
type MailingListMemberModel struct {
	// Required fields
	ListAddress   types.String `tfsdk:"list_address"`   // The mailing list address
	MemberAddress types.String `tfsdk:"member_address"` // The member's email address

	// Optional fields
	Name       types.String `tfsdk:"name"`
	Subscribed types.Bool   `tfsdk:"subscribed"`
	Vars       types.Map    `tfsdk:"vars"` // Map of string to string for variables

	// Computed fields
	Id types.String `tfsdk:"id"` // Composite ID: list_address/member_address
}
