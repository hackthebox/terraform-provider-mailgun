// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ListDataSource{}
	_ datasource.DataSourceWithConfigure = &ListDataSource{}
)

// DataSource is the data source implementation.
type ListDataSource struct {
	client *mailgun.Client
}

// Metadata returns the data source type name.
func (d *ListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

// Schema defines the schema for the data source.
func (d *ListDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainsListDataSourceSchema()
}

// Configure adds the provider-configured client to the data source.
func (d *ListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *ListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If a client is not configured, return
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	// Set default values for optional parameters
	limit := int64(100)
	if !data.Limit.IsNull() {
		limit = data.Limit.ValueInt64()
	}

	// Create list options (v5 API)
	opts := &mailgun.ListDomainsOptions{
		Limit: int(limit),
	}

	// Get domains from Mailgun API
	domainsIterator := d.client.ListDomains(opts)

	var domains []mtypes.Domain
	var domainItems []ItemsValue

	// Collect domains
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// Get domains
	var page []mtypes.Domain
	for domainsIterator.Next(ctx, &page) {
		domains = append(domains, page...)
	}

	// Check for errors
	if err := domainsIterator.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Domains",
			fmt.Sprintf("Unable to list domains: %s", err),
		)
		return
	}

	// Convert domains to Terraform model
	for _, domain := range domains {
		// Create null disabled object (disabled details not available in list response)
		disabledObjValue := types.ObjectNull(map[string]attr.Type{
			"code":        types.StringType,
			"note":        types.StringType,
			"permanently": types.BoolType,
			"reason":      types.StringType,
			"until":       types.StringType,
		})

		// Format created at time
		createdAt := domain.CreatedAt.String()

		// Create domain item
		item := ItemsValue{
			CreatedAt:                  types.StringValue(createdAt),
			Disabled:                   disabledObjValue,
			Id:                         types.StringValue(domain.Name), // Using name as ID
			IsDisabled:                 types.BoolValue(false),         // Default value
			Name:                       types.StringValue(domain.Name),
			RequireTls:                 types.BoolValue(false), // Default value
			SkipVerification:           types.BoolValue(false), // Default value
			SmtpLogin:                  types.StringValue(domain.SMTPLogin),
			SmtpPassword:               types.StringValue(domain.SMTPPassword),
			SpamAction:                 types.StringValue(string(domain.SpamAction)),
			State:                      types.StringValue(domain.State),
			TrackingHost:               types.StringValue(""),  // Default value
			ItemsType:                  types.StringValue(""),  // Default value
			UseAutomaticSenderSecurity: types.BoolValue(false), // Default value
			WebPrefix:                  types.StringValue(""),  // Default value
			WebScheme:                  types.StringValue(domain.WebScheme),
			Wildcard:                   types.BoolValue(domain.Wildcard),
			state:                      attr.ValueStateKnown,
		}

		domainItems = append(domainItems, item)
	}

	// Convert domain items to List
	itemsList, diags := convertItemsToList(ctx, domainItems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set values in the data model
	data.Items = itemsList
	data.TotalCount = types.Int64Value(int64(len(domains)))

	// Set default values for optional parameters that weren't set
	if data.Authority.IsNull() {
		data.Authority = types.StringValue("")
	}
	if data.IncludeSubaccounts.IsNull() {
		data.IncludeSubaccounts = types.BoolValue(false)
	}
	if data.Limit.IsNull() {
		data.Limit = types.Int64Value(limit)
	}
	if data.Search.IsNull() {
		data.Search = types.StringValue("")
	}
	if data.Skip.IsNull() {
		data.Skip = types.Int64Value(0)
	}
	if data.Sort.IsNull() {
		data.Sort = types.StringValue("")
	}
	if data.State.IsNull() {
		data.State = types.StringValue("")
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to convert ItemsValue slice to types.List.
func convertItemsToList(ctx context.Context, items []ItemsValue) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	// If no items, return empty list
	if len(items) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: ItemsValue{}.AttributeTypes(ctx),
		}), diags
	}

	// Convert each item to ObjectValue
	objectValues := make([]attr.Value, 0, len(items))
	for _, item := range items {
		objValue, objDiags := item.ToObjectValue(ctx)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{
				AttrTypes: ItemsValue{}.AttributeTypes(ctx),
			}), diags
		}
		objectValues = append(objectValues, objValue)
	}

	// Create list from object values
	listValue, listDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: ItemsValue{}.AttributeTypes(ctx),
		},
		objectValues,
	)
	diags.Append(listDiags...)

	return listValue, diags
}
