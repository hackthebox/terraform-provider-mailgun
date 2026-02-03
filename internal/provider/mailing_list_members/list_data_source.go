// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ datasource.DataSource              = &mailingListMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &mailingListMembersDataSource{}
)

// NewMailingListMembersDataSource creates a new mailing list members data source.
func NewMailingListMembersDataSource() datasource.DataSource {
	return &mailingListMembersDataSource{}
}

type mailingListMembersDataSource struct {
	client *mailgun.Client
}

func (d *mailingListMembersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailing_list_members"
}

func (d *mailingListMembersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = MailingListMembersDataSourceSchema()
}

func (d *mailingListMembersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *mailingListMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MailingListMembersDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listAddress := state.ListAddress.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// List all members using the iterator
	iterator := d.client.ListMembers(listAddress, nil)

	var allMembers []mtypes.Member
	var page []mtypes.Member

	// Get first page
	if iterator.First(readCtx, &page) {
		allMembers = append(allMembers, page...)

		// Get remaining pages
		for iterator.Next(readCtx, &page) {
			allMembers = append(allMembers, page...)
		}
	}

	if err := iterator.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Mailgun Mailing List Members",
			fmt.Sprintf("Could not list members for list %s: %s", listAddress, err.Error()),
		)
		return
	}

	// Map to state
	state.TotalCount = types.Int64Value(int64(len(allMembers)))
	state.Members = make([]MailingListMemberItemModel, len(allMembers))

	for i, member := range allMembers {
		state.Members[i] = MailingListMemberItemModel{
			Address: types.StringValue(member.Address),
		}

		if member.Name != "" {
			state.Members[i].Name = types.StringValue(member.Name)
		} else {
			state.Members[i].Name = types.StringNull()
		}

		if member.Subscribed != nil {
			state.Members[i].Subscribed = types.BoolValue(*member.Subscribed)
		} else {
			state.Members[i].Subscribed = types.BoolValue(true)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
