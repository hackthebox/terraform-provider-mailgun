// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists

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
	_ datasource.DataSource              = &mailingListsDataSource{}
	_ datasource.DataSourceWithConfigure = &mailingListsDataSource{}
)

// NewMailingListsDataSource creates a new mailing lists data source.
func NewMailingListsDataSource() datasource.DataSource {
	return &mailingListsDataSource{}
}

type mailingListsDataSource struct {
	client *mailgun.Client
}

func (d *mailingListsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailing_lists"
}

func (d *mailingListsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = MailingListsDataSourceSchema()
}

func (d *mailingListsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mailingListsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MailingListsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// List all mailing lists using the iterator
	iterator := d.client.ListMailingLists(nil)

	var allLists []mtypes.MailingList
	var page []mtypes.MailingList

	// Get first page
	if iterator.First(readCtx, &page) {
		allLists = append(allLists, page...)

		// Get remaining pages
		for iterator.Next(readCtx, &page) {
			allLists = append(allLists, page...)
		}
	}

	if err := iterator.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Mailgun Mailing Lists",
			fmt.Sprintf("Could not list mailing lists: %s", err.Error()),
		)
		return
	}

	// Map to state
	state.TotalCount = types.Int64Value(int64(len(allLists)))
	state.MailingLists = make([]MailingListItemModel, len(allLists))

	for i, list := range allLists {
		state.MailingLists[i] = MailingListItemModel{
			Address:      types.StringValue(list.Address),
			AccessLevel:  types.StringValue(string(list.AccessLevel)),
			CreatedAt:    types.StringValue(list.CreatedAt.String()),
			MembersCount: types.Int64Value(int64(list.MembersCount)),
		}

		if list.Name != "" {
			state.MailingLists[i].Name = types.StringValue(list.Name)
		} else {
			state.MailingLists[i].Name = types.StringNull()
		}

		if list.Description != "" {
			state.MailingLists[i].Description = types.StringValue(list.Description)
		} else {
			state.MailingLists[i].Description = types.StringNull()
		}

		if list.ReplyPreference != "" {
			state.MailingLists[i].ReplyPreference = types.StringValue(string(list.ReplyPreference))
		} else {
			state.MailingLists[i].ReplyPreference = types.StringNull()
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
