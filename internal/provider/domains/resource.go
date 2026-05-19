// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithConfigure   = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

// DomainResource is the resource implementation.
type DomainResource struct {
	client *mailgun.Client
}

// Metadata returns the resource type name.
func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the resource.
func (r *DomainResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainResourceSchema()
}

// Configure adds the provider-configured client to the resource.
func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new resource.
func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	// Validate required fields
	if plan.Name.IsNull() || plan.Name.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Domain Name",
			"The domain name is required to create a domain.",
		)
		return
	}

	domainName := plan.Name.ValueString()

	// Build CreateDomainOptions
	opts := &mailgun.CreateDomainOptions{}

	if !plan.SmtpPassword.IsNull() {
		opts.Password = plan.SmtpPassword.ValueString()
	}

	if !plan.SpamAction.IsNull() {
		opts.SpamAction = mtypes.SpamAction(plan.SpamAction.ValueString())
	}

	if !plan.Wildcard.IsNull() {
		opts.Wildcard = plan.Wildcard.ValueBool()
	}

	if !plan.ForceDkimAuthority.IsNull() {
		opts.ForceDKIMAuthority = plan.ForceDkimAuthority.ValueBool()
	}

	if !plan.DkimKeySize.IsNull() {
		// DkimKeySize is stored as string in the model, convert to int
		var keySize int
		switch plan.DkimKeySize.ValueString() {
		case "1024":
			keySize = 1024
		case "2048":
			keySize = 2048
		}
		if keySize != 0 {
			opts.DKIMKeySize = keySize
		}
	}

	if !plan.UseAutomaticSenderSecurity.IsNull() {
		opts.UseAutomaticSenderSecurity = plan.UseAutomaticSenderSecurity.ValueBool()
	}

	if !plan.WebScheme.IsNull() {
		opts.WebScheme = plan.WebScheme.ValueString()
	}

	if !plan.Ips.IsNull() && plan.Ips.ValueString() != "" {
		// IPs is a comma-separated string in Terraform, SDK expects []string
		opts.IPs = strings.Split(plan.Ips.ValueString(), ",")
	}

	// Note: The following fields are supported by the API but not yet by the SDK:
	// - WebPrefix, DkimHostName, DkimSelector, ForceRootDkimHost,
	//   EncryptIncomingMessage, PoolId
	// See: https://github.com/mailgun/mailgun-go TODO(DE-1599)

	// Create context with timeout
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create the domain via Mailgun API
	// CreateDomain returns GetDomainResponse which includes all domain details
	domainResp, err := r.client.CreateDomain(createCtx, domainName, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Domain",
			fmt.Sprintf("Could not create domain %s: %s", domainName, err),
		)
		return
	}

	// Map response to a plan model
	plan = mapDomainResponseToModel(domainResp, plan)

	// Fetch authentication DNS records (DMARC) separately — not yet in SDK
	authRecords, err := getAuthenticationDNSRecords(ctx, r.client, domainName)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Could Not Fetch Authentication DNS Records",
			fmt.Sprintf("Domain %s was created, but authentication DNS records could not be retrieved: %s", domainName, err),
		)
	}
	setAuthenticationDNSRecords(ctx, authRecords, &plan)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	// Get domain name from state
	domainName := state.Name.ValueString()
	if domainName == "" {
		resp.Diagnostics.AddError(
			"Missing Domain Name",
			"The domain name is required to read a domain.",
		)
		return
	}

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get the domain via Mailgun API
	domainResp, err := r.client.GetDomain(readCtx, domainName, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain",
			fmt.Sprintf("Could not read domain %s: %s", domainName, err),
		)
		return
	}

	// Map response to a state model
	state = mapDomainResponseToModel(domainResp, state)

	// Fetch authentication DNS records (DMARC) separately — not yet in SDK
	authRecords, err := getAuthenticationDNSRecords(ctx, r.client, domainName)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Could Not Fetch Authentication DNS Records",
			fmt.Sprintf("Authentication DNS records could not be retrieved for domain %s: %s", domainName, err),
		)
	}
	setAuthenticationDNSRecords(ctx, authRecords, &state)

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing resource.
func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainModel
	var state DomainModel

	// Read Terraform plan and current state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	domainName := plan.Name.ValueString()

	// Build update options for fields that can be updated via SDK
	// Note: SDK only supports WebScheme, WebPrefix, RequireTLS, SkipVerification, UseAutomaticSenderSecurity
	opts := &mailgun.UpdateDomainOptions{}
	hasChanges := false

	if !plan.WebScheme.Equal(state.WebScheme) && !plan.WebScheme.IsNull() {
		opts.WebScheme = plan.WebScheme.ValueString()
		hasChanges = true
	}

	if !plan.WebPrefix.Equal(state.WebPrefix) && !plan.WebPrefix.IsNull() {
		opts.WebPrefix = plan.WebPrefix.ValueString()
		hasChanges = true
	}

	if !plan.UseAutomaticSenderSecurity.Equal(state.UseAutomaticSenderSecurity) && !plan.UseAutomaticSenderSecurity.IsNull() {
		val := plan.UseAutomaticSenderSecurity.ValueBool()
		opts.UseAutomaticSenderSecurity = &val
		hasChanges = true
	}

	// Create context with timeout
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Apply updates if there are changes
	if hasChanges {
		err := r.client.UpdateDomain(updateCtx, domainName, opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Domain",
				fmt.Sprintf("Could not update domain %s: %s", domainName, err),
			)
			return
		}
	}

	// Fetch the latest domain state
	domainResp, err := r.client.GetDomain(updateCtx, domainName, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain After Update",
			fmt.Sprintf("Could not read domain %s: %s", domainName, err),
		)
		return
	}

	// Map response to model
	plan = mapDomainResponseToModel(domainResp, plan)

	// Fetch authentication DNS records (DMARC) separately — not yet in SDK
	authRecords, err := getAuthenticationDNSRecords(ctx, r.client, domainName)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Could Not Fetch Authentication DNS Records",
			fmt.Sprintf("Authentication DNS records could not be retrieved for domain %s: %s", domainName, err),
		)
	}
	setAuthenticationDNSRecords(ctx, authRecords, &plan)

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes an existing resource.
func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	domainName := state.Name.ValueString()
	if domainName == "" {
		resp.Diagnostics.AddError(
			"Missing Domain Name",
			"The domain name is required to delete a domain.",
		)
		return
	}

	// Create context with timeout
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Delete the domain via Mailgun API
	err := r.client.DeleteDomain(deleteCtx, domainName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Domain",
			fmt.Sprintf("Could not delete domain %s: %s", domainName, err),
		)
		return
	}

	// State is automatically removed by Terraform after successful deletion
}

// ImportState imports an existing domain by name
func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID passed is the domain name
	domainName := req.ID

	// Validate that client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Mailgun Client",
			"The Mailgun client has not been properly configured. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	// Create context with timeout
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get the domain via Mailgun API
	domainResp, err := r.client.GetDomain(importCtx, domainName, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Domain",
			fmt.Sprintf("Could not import domain %s: %s", domainName, err),
		)
		return
	}

	// Create a new model with imported values
	var state DomainModel
	state.Name = types.StringValue(domainName)
	state = mapDomainResponseToModel(domainResp, state)

	// Fetch authentication DNS records (DMARC) separately — not yet in SDK
	authRecords, err := getAuthenticationDNSRecords(importCtx, r.client, domainName)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Could Not Fetch Authentication DNS Records",
			fmt.Sprintf("Authentication DNS records could not be retrieved for domain %s: %s", domainName, err),
		)
	}
	setAuthenticationDNSRecords(importCtx, authRecords, &state)

	// Save imported state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// setAuthenticationDNSRecords maps a slice of mtypes.DNSRecord into the model's
// AuthenticationDnsRecords list attribute.
func setAuthenticationDNSRecords(ctx context.Context, records []mtypes.DNSRecord, model *DomainModel) {
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

// dmarcRecordResponse is the response from the Mailgun DMARC records API.
// https://documentation.mailgun.com/docs/validate/oas/openapi-final/dmarc-reports/get-v1-dmarc-records-domain-
type dmarcRecordResponse struct {
	Entry      string `json:"entry"`
	Current    string `json:"current"`
	Configured bool   `json:"configured"`
}

// getAuthenticationDNSRecords fetches the Mailgun-generated DMARC record for a domain
// via GET /v1/dmarc/records/{domain} and returns it as a DNSRecord ready for use in CF.
func getAuthenticationDNSRecords(ctx context.Context, client *mailgun.Client, domainName string) ([]mtypes.DNSRecord, error) {
	url := fmt.Sprintf("%s/v1/dmarc/records/%s", client.APIBase(), domainName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("api", client.APIKey())

	resp, err := client.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dmarcRecordResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Entry == "" {
		return nil, nil
	}

	valid := "unknown"
	if result.Configured {
		valid = "valid"
	}

	return []mtypes.DNSRecord{
		{
			RecordType: "TXT",
			Name:       "_dmarc." + domainName,
			Value:      result.Entry,
			Valid:      valid,
		},
	}, nil
}

// mapDomainResponseToModel maps Mailgun API response to Terraform model
func mapDomainResponseToModel(domainResp mtypes.GetDomainResponse, model DomainModel) DomainModel {
	ctx := context.Background()

	// Set computed attributes from domainResp.Domain
	// Use domain name as Terraform resource ID (Mailgun API uses domain names for lookups)
	model.Id = types.StringValue(domainResp.Domain.Name)
	model.Name = types.StringValue(domainResp.Domain.Name)
	model.CreatedAt = types.StringValue(domainResp.Domain.CreatedAt.String())
	model.State = types.StringValue(domainResp.Domain.State)
	model.SmtpLogin = types.StringValue(domainResp.Domain.SMTPLogin)
	model.IsDisabled = types.BoolValue(domainResp.Domain.IsDisabled)
	model.RequireTls = types.BoolValue(domainResp.Domain.RequireTLS)
	model.SkipVerification = types.BoolValue(domainResp.Domain.SkipVerification)
	model.DomainType = types.StringValue(domainResp.Domain.Type)
	model.TrackingHost = types.StringValue(domainResp.Domain.TrackingHost)

	// Set Optional/Computed fields from response
	model.SpamAction = types.StringValue(string(domainResp.Domain.SpamAction))
	model.Wildcard = types.BoolValue(domainResp.Domain.Wildcard)
	model.WebScheme = types.StringValue(domainResp.Domain.WebScheme)
	model.WebPrefix = types.StringValue(domainResp.Domain.WebPrefix)
	model.UseAutomaticSenderSecurity = types.BoolValue(domainResp.Domain.UseAutomaticSenderSecurity)

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
		model.ReceivingDnsRecords = receivingList
	} else if model.ReceivingDnsRecords.IsUnknown() {
		receivingDnsRecordsType := ReceivingDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: ReceivingDnsRecordsValue{}.AttributeTypes(ctx)},
		}
		model.ReceivingDnsRecords = types.ListNull(receivingDnsRecordsType)
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
		model.SendingDnsRecords = sendingList
	} else if model.SendingDnsRecords.IsUnknown() {
		sendingDnsRecordsType := SendingDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: SendingDnsRecordsValue{}.AttributeTypes(ctx)},
		}
		model.SendingDnsRecords = types.ListNull(sendingDnsRecordsType)
	}

	// Authentication DNS records are populated separately via getAuthenticationDNSRecords;
	// initialize as null here so state is always well-defined.
	if model.AuthenticationDnsRecords.IsUnknown() {
		authDnsRecordsType := AuthenticationDnsRecordsType{
			ObjectType: types.ObjectType{AttrTypes: AuthenticationDnsRecordsValue{}.AttributeTypes(ctx)},
		}
		model.AuthenticationDnsRecords = types.ListNull(authDnsRecordsType)
	}

	// Set request-only fields to null if unknown (not returned in response)
	if model.DkimHostName.IsUnknown() {
		model.DkimHostName = types.StringNull()
	}
	if model.DkimKeySize.IsUnknown() {
		model.DkimKeySize = types.StringNull()
	}
	if model.DkimSelector.IsUnknown() {
		model.DkimSelector = types.StringNull()
	}
	if model.EncryptIncomingMessage.IsUnknown() {
		model.EncryptIncomingMessage = types.BoolNull()
	}
	if model.ForceDkimAuthority.IsUnknown() {
		model.ForceDkimAuthority = types.BoolNull()
	}
	if model.ForceRootDkimHost.IsUnknown() {
		model.ForceRootDkimHost = types.BoolNull()
	}
	if model.Ips.IsUnknown() {
		model.Ips = types.StringNull()
	}
	if model.PoolId.IsUnknown() {
		model.PoolId = types.StringNull()
	}
	if model.SmtpPassword.IsUnknown() {
		model.SmtpPassword = types.StringNull()
	}

	return model
}
