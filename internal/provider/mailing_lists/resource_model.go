// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MailingListModel represents the Terraform state for a Mailgun mailing list.
type MailingListModel struct {
	// Required fields
	Address types.String `tfsdk:"address"` // Full email address of the list (e.g., list@example.com)

	// Optional fields
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	AccessLevel     types.String `tfsdk:"access_level"`
	ReplyPreference types.String `tfsdk:"reply_preference"`

	// Computed fields
	Id           types.String `tfsdk:"id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	MembersCount types.Int64  `tfsdk:"members_count"`
}
