// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MailingListMembersDataSourceModel represents the data source state for listing mailing list members.
type MailingListMembersDataSourceModel struct {
	ListAddress types.String                 `tfsdk:"list_address"`
	TotalCount  types.Int64                  `tfsdk:"total_count"`
	Members     []MailingListMemberItemModel `tfsdk:"members"`
}

// MailingListMemberItemModel represents a single member in the list.
type MailingListMemberItemModel struct {
	Address    types.String `tfsdk:"address"`
	Name       types.String `tfsdk:"name"`
	Subscribed types.Bool   `tfsdk:"subscribed"`
}
