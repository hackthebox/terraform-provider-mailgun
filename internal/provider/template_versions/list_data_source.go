// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions

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
	_ datasource.DataSource              = &templateVersionsDataSource{}
	_ datasource.DataSourceWithConfigure = &templateVersionsDataSource{}
)

// NewTemplateVersionsDataSource creates a new template versions list data source.
func NewTemplateVersionsDataSource() datasource.DataSource {
	return &templateVersionsDataSource{}
}

type templateVersionsDataSource struct {
	client *mailgun.Client
}

func (d *templateVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_versions"
}

func (d *templateVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = TemplateVersionsDataSourceSchema()
}

func (d *templateVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *templateVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TemplateVersionsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	templateName := state.TemplateName.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// List all versions using the iterator
	iterator := d.client.ListTemplateVersions(domain, templateName, nil)

	var allVersions []mtypes.TemplateVersion
	var page []mtypes.TemplateVersion

	// Get first page
	if iterator.First(readCtx, &page) {
		allVersions = append(allVersions, page...)

		// Get remaining pages
		for iterator.Next(readCtx, &page) {
			allVersions = append(allVersions, page...)
		}
	}

	if err := iterator.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Mailgun Template Versions",
			fmt.Sprintf("Could not list template versions for %s/%s: %s", domain, templateName, err.Error()),
		)
		return
	}

	// Map to state
	state.TotalCount = types.Int64Value(int64(len(allVersions)))
	state.Versions = make([]TemplateVersionItemModel, len(allVersions))

	for i, v := range allVersions {
		state.Versions[i] = TemplateVersionItemModel{
			Tag:       types.StringValue(v.Tag),
			Engine:    types.StringValue(string(v.Engine)),
			Active:    types.BoolValue(v.Active),
			CreatedAt: types.StringValue(v.CreatedAt.String()),
		}

		if v.Comment != "" {
			state.Versions[i].Comment = types.StringValue(v.Comment)
		} else {
			state.Versions[i].Comment = types.StringNull()
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
