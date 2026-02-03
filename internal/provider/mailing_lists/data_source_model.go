// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MailingListsDataSourceModel represents the data source state for listing mailing lists.
type MailingListsDataSourceModel struct {
	TotalCount   types.Int64            `tfsdk:"total_count"`
	MailingLists []MailingListItemModel `tfsdk:"mailing_lists"`
}

// MailingListItemModel represents a single mailing list in the list.
type MailingListItemModel struct {
	Address         types.String `tfsdk:"address"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	AccessLevel     types.String `tfsdk:"access_level"`
	ReplyPreference types.String `tfsdk:"reply_preference"`
	CreatedAt       types.String `tfsdk:"created_at"`
	MembersCount    types.Int64  `tfsdk:"members_count"`
}
