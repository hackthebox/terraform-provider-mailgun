// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DomainDataSource{}
	_ datasource.DataSourceWithConfigure = &DomainDataSource{}
)

// setAuthenticationDNSRecordsOnDataSource maps a slice of mtypes.DNSRecord into the
// data source model's AuthenticationDnsRecords list attribute.
func setAuthenticationDNSRecordsOnDataSource(ctx context.Context, records []mtypes.DNSRecord, model *DomainDataSourceModel) {
	authType := AuthenticationDnsRecordsType{
		ObjectType: types.ObjectType{AttrTypes: AuthenticationDnsRecordsValue{}.AttributeTypes(ctx)},
	}
	if len(records) == 0 {
		model.AuthenticationDnsRecords = types.ListNull(authType)
		return
	}

	authRecords := make([]AuthenticationDnsRecordsValue, len(records))
	for i, record := range records {
		cachedList, _ := types.ListValueFrom(ctx, types.StringType, record.Cached)
		authRecords[i] = AuthenticationDnsRecordsValue{
			Cached:     cachedList,
			IsActive:   types.BoolValue(record.Active),
			Name:       types.StringValue(record.Name),
			Priority:   types.StringValue(record.Priority),
			RecordType: types.StringValue(record.RecordType),
			Valid:      types.StringValue(record.Valid),
			Value:      types.StringValue(record.Value),
			state:      attr.ValueStateKnown,
		}
	}

	authList, _ := types.ListValueFrom(ctx, authType, authRecords)
	model.AuthenticationDnsRecords = authList
}

// DomainDataSource is the single domain data source implementation.
type DomainDataSource struct {
	client *mailgun.Client
}

// Metadata returns the data source type name.
func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the data source.
func (d *DomainDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainDataSourceSchema()
}

// Configure adds the provider-configured client to the data source.
func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDataSourceModel

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

	domainName := data.Name.ValueString()
	if domainName == "" {
		resp.Diagnostics.AddError(
			"Missing Domain Name",
			"The domain name is required to lookup a domain.",
		)
		return
	}

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get the domain via Mailgun API
	domainResp, err := d.client.GetDomain(readCtx, domainName, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain",
			fmt.Sprintf("Could not read domain %s: %s", domainName, err),
		)
		return
	}

	// Map response to data source model
	data.CreatedAt = types.StringValue(domainResp.Domain.CreatedAt.String())
	data.Id = types.StringValue(domainResp.Domain.ID)
	data.IsDisabled = types.BoolValue(domainResp.Domain.IsDisabled)
	data.RequireTls = types.BoolValue(domainResp.Domain.RequireTLS)
	data.SkipVerification = types.BoolValue(domainResp.Domain.SkipVerification)
	data.SmtpLogin = types.StringValue(domainResp.Domain.SMTPLogin)
	data.SmtpPassword = types.StringValue(domainResp.Domain.SMTPPassword)
	data.SpamAction = types.StringValue(string(domainResp.Domain.SpamAction))
	data.State = types.StringValue(domainResp.Domain.State)
	data.TrackingHost = types.StringValue(domainResp.Domain.TrackingHost)
	data.DomainType = types.StringValue(domainResp.Domain.Type)
	data.UseAutomaticSenderSecurity = types.BoolValue(domainResp.Domain.UseAutomaticSenderSecurity)
	data.WebPrefix = types.StringValue(domainResp.Domain.WebPrefix)
	data.WebScheme = types.StringValue(domainResp.Domain.WebScheme)
	data.Wildcard = types.BoolValue(domainResp.Domain.Wildcard)

	// Map DNS records from response
	if len(domainResp.ReceivingDNSRecords) > 0 {
		receivingRecords := make([]ReceivingDnsRecordsValue, len(domainResp.ReceivingDNSRecords))
		for i, record := range domainResp.ReceivingDNSRecords {
			cachedList, _ := types.ListValueFrom(ctx, types.StringType, record.Cached)
			receivingRecords[i] = ReceivingDnsRecordsValue{
				Cached:     cachedList,
				IsActive:   types.BoolValue(record.Active),
				Name:       types.StringValue(record.Name),
				Priority:   types.StringValue(record.Priority),
				RecordType: types.StringValue(record.RecordType),
				Valid:      types.StringValue(record.Valid),
				Value:      types.StringValue(record.Value),
				state:      attr.ValueStateKnown,
			}
		}
		receivingList, _ := types.ListValueFrom(ctx, ReceivingDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: ReceivingDnsRecordsValue{}.AttributeTypes(ctx)},
		}, receivingRecords)
		data.ReceivingDnsRecords = receivingList
	} else {
		receivingDnsRecordsType := ReceivingDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: ReceivingDnsRecordsValue{}.AttributeTypes(ctx)},
		}
		data.ReceivingDnsRecords = types.ListNull(receivingDnsRecordsType)
	}

	if len(domainResp.SendingDNSRecords) > 0 {
		sendingRecords := make([]SendingDnsRecordsValue, len(domainResp.SendingDNSRecords))
		for i, record := range domainResp.SendingDNSRecords {
			cachedList, _ := types.ListValueFrom(ctx, types.StringType, record.Cached)
			sendingRecords[i] = SendingDnsRecordsValue{
				Cached:     cachedList,
				IsActive:   types.BoolValue(record.Active),
				Name:       types.StringValue(record.Name),
				Priority:   types.StringValue(record.Priority),
				RecordType: types.StringValue(record.RecordType),
				Valid:      types.StringValue(record.Valid),
				Value:      types.StringValue(record.Value),
				state:      attr.ValueStateKnown,
			}
		}
		sendingList, _ := types.ListValueFrom(ctx, SendingDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: SendingDnsRecordsValue{}.AttributeTypes(ctx)},
		}, sendingRecords)
		data.SendingDnsRecords = sendingList
	} else {
		sendingDnsRecordsType := SendingDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: SendingDnsRecordsValue{}.AttributeTypes(ctx)},
		}
		data.SendingDnsRecords = types.ListNull(sendingDnsRecordsType)
	}

	// Fetch authentication DNS records (DMARC) separately — not yet in SDK
	authRecords, authErr := getAuthenticationDNSRecords(readCtx, d.client, domainName)
	if authErr != nil {
		resp.Diagnostics.AddWarning(
			"Could Not Fetch Authentication DNS Records",
			fmt.Sprintf("Authentication DNS records could not be retrieved for domain %s: %s", domainName, authErr),
		)
	}
	setAuthenticationDNSRecordsOnDataSource(readCtx, authRecords, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
