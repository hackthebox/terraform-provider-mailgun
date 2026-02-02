// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates

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
	_ datasource.DataSource              = &templatesDataSource{}
	_ datasource.DataSourceWithConfigure = &templatesDataSource{}
)

// NewTemplatesDataSource creates a new templates list data source.
func NewTemplatesDataSource() datasource.DataSource {
	return &templatesDataSource{}
}

type templatesDataSource struct {
	client *mailgun.Client
}

func (d *templatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates"
}

func (d *templatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = TemplatesDataSourceSchema()
}

func (d *templatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *templatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TemplatesDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// List all templates using the iterator
	iterator := d.client.ListTemplates(domain, nil)

	var allTemplates []mtypes.Template
	var page []mtypes.Template

	// Get first page
	if iterator.First(readCtx, &page) {
		allTemplates = append(allTemplates, page...)

		// Get remaining pages
		for iterator.Next(readCtx, &page) {
			allTemplates = append(allTemplates, page...)
		}
	}

	if err := iterator.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Mailgun Templates",
			fmt.Sprintf("Could not list templates for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Map to state
	state.TotalCount = types.Int64Value(int64(len(allTemplates)))
	state.Templates = make([]TemplateItemModel, len(allTemplates))

	for i, t := range allTemplates {
		state.Templates[i] = TemplateItemModel{
			Name:        types.StringValue(t.Name),
			Description: types.StringValue(t.Description),
			CreatedAt:   types.StringValue(t.CreatedAt.String()),
		}

		if t.Version.Tag != "" {
			state.Templates[i].VersionTag = types.StringValue(t.Version.Tag)
			state.Templates[i].VersionEngine = types.StringValue(string(t.Version.Engine))
			state.Templates[i].VersionActive = types.BoolValue(t.Version.Active)
		} else {
			state.Templates[i].VersionTag = types.StringNull()
			state.Templates[i].VersionEngine = types.StringNull()
			state.Templates[i].VersionActive = types.BoolNull()
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
