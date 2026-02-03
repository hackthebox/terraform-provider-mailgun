// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ resource.Resource                = &domainDkimKeyResource{}
	_ resource.ResourceWithConfigure   = &domainDkimKeyResource{}
	_ resource.ResourceWithImportState = &domainDkimKeyResource{}
)

// NewDomainDkimKeyResource creates a new domain DKIM key resource.
func NewDomainDkimKeyResource() resource.Resource {
	return &domainDkimKeyResource{}
}

type domainDkimKeyResource struct {
	client *mailgun.Client
}

func (r *domainDkimKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dkim_key"
}

func (r *domainDkimKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainDkimKeyResourceSchema()
}

func (r *domainDkimKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailgun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mailgun.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *domainDkimKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainDkimKeyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	selector := plan.Selector.ValueString()

	// Create context with timeout
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build create options
	opts := &mailgun.CreateDomainKeyOptions{}
	if !plan.Bits.IsNull() && !plan.Bits.IsUnknown() {
		opts.Bits = int(plan.Bits.ValueInt64())
	}

	// Create the DKIM key
	domainKey, err := r.client.CreateDomainKey(createCtx, domain, selector, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Domain DKIM Key",
			fmt.Sprintf("Could not create DKIM key for domain %s with selector %s: %s", domain, selector, err.Error()),
		)
		return
	}

	// Map to state first
	mapDomainKeyToModel(&plan, domain, selector, domainKey)

	// Set active state from API response (newly created keys are inactive by default)
	plan.Active = types.BoolValue(domainKey.DNSRecord.Active)

	// If user explicitly requested active = true, activate the key
	var planActive types.Bool
	diags = req.Plan.GetAttribute(ctx, path.Root("active"), &planActive)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !planActive.IsNull() && planActive.ValueBool() {
		activateCtx, cancelActivate := context.WithTimeout(ctx, 30*time.Second)
		defer cancelActivate()

		if err := r.client.ActivateDomainKey(activateCtx, domain, selector); err != nil {
			resp.Diagnostics.AddError(
				"Error Activating Domain DKIM Key",
				fmt.Sprintf("Created DKIM key but could not activate: %s", err.Error()),
			)
			return
		}
		plan.Active = types.BoolValue(true)
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainDkimKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainDkimKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	selector := state.Selector.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List domain keys and find ours
	domainKey, found, err := r.findDomainKey(readCtx, domain, selector)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain DKIM Key",
			fmt.Sprintf("Could not read DKIM key for domain %s with selector %s: %s", domain, selector, err.Error()),
		)
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map to state (preserve bits as it's not returned by the API)
	mapDomainKeyToModel(&state, domain, selector, domainKey)

	// Determine active state from DNS record
	state.Active = types.BoolValue(domainKey.DNSRecord.Active)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *domainDkimKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainDkimKeyModel
	var state DomainDkimKeyModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	selector := plan.Selector.ValueString()

	// Only active can be updated
	if !plan.Active.Equal(state.Active) {
		updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if plan.Active.ValueBool() {
			if err := r.client.ActivateDomainKey(updateCtx, domain, selector); err != nil {
				resp.Diagnostics.AddError(
					"Error Activating Domain DKIM Key",
					fmt.Sprintf("Could not activate DKIM key: %s", err.Error()),
				)
				return
			}
		} else {
			if err := r.client.DeactivateDomainKey(updateCtx, domain, selector); err != nil {
				resp.Diagnostics.AddError(
					"Error Deactivating Domain DKIM Key",
					fmt.Sprintf("Could not deactivate DKIM key: %s", err.Error()),
				)
				return
			}
		}
	}

	// Read back the current state
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	domainKey, found, err := r.findDomainKey(readCtx, domain, selector)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain DKIM Key After Update",
			fmt.Sprintf("Could not read DKIM key after update: %s", err.Error()),
		)
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Domain DKIM Key Not Found After Update",
			"The DKIM key was not found after the update operation.",
		)
		return
	}

	// Map to state
	mapDomainKeyToModel(&plan, domain, selector, domainKey)
	plan.Active = types.BoolValue(domainKey.DNSRecord.Active)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainDkimKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainDkimKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	selector := state.Selector.ValueString()

	// Create context with timeout
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.client.DeleteDomainKey(deleteCtx, domain, selector); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Domain DKIM Key",
			fmt.Sprintf("Could not delete DKIM key for domain %s with selector %s: %s", domain, selector, err.Error()),
		)
		return
	}
}

func (r *domainDkimKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: domain/selector
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format: domain/selector",
		)
		return
	}

	domain := parts[0]
	selector := parts[1]

	// Read the key
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	domainKey, found, err := r.findDomainKey(readCtx, domain, selector)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Domain DKIM Key",
			fmt.Sprintf("Could not read DKIM key for domain %s with selector %s: %s", domain, selector, err.Error()),
		)
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Domain DKIM Key Not Found",
			fmt.Sprintf("DKIM key with selector %s not found for domain %s", selector, domain),
		)
		return
	}

	var state DomainDkimKeyModel
	mapDomainKeyToModel(&state, domain, selector, domainKey)
	state.Active = types.BoolValue(domainKey.DNSRecord.Active)
	// Bits is not returned by the API, set to default
	state.Bits = types.Int64Value(1024)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// findDomainKey searches for a domain key with the given selector.
func (r *domainDkimKeyResource) findDomainKey(ctx context.Context, domain, selector string) (mtypes.DomainKey, bool, error) {
	iter := r.client.ListDomainKeys(domain)

	// Use First() instead of Next() because the SDK's DomainKeysIterator.Next()
	// returns false even on first call due to empty Paging.Next from the API.
	// A domain can have at most 5 DKIM keys, so pagination isn't needed.
	var keys []mtypes.DomainKey
	if !iter.First(ctx, &keys) {
		if err := iter.Err(); err != nil {
			return mtypes.DomainKey{}, false, err
		}
		// No error but First returned false - shouldn't happen, but handle gracefully
		return mtypes.DomainKey{}, false, nil
	}

	for _, key := range keys {
		if key.Selector == selector {
			return key, true, nil
		}
	}

	return mtypes.DomainKey{}, false, nil
}

// mapDomainKeyToModel maps a Mailgun domain key to the Terraform model.
func mapDomainKeyToModel(model *DomainDkimKeyModel, domain, selector string, key mtypes.DomainKey) {
	model.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, selector))
	model.Domain = types.StringValue(domain)
	model.Selector = types.StringValue(selector)
	model.SigningDomain = types.StringValue(key.SigningDomain)

	// Map DNS record details
	model.DnsRecordName = types.StringValue(key.DNSRecord.Name)
	model.DnsRecordType = types.StringValue(key.DNSRecord.RecordType)
	model.DnsRecordValue = types.StringValue(key.DNSRecord.Value)
	model.DnsRecordValid = types.StringValue(key.DNSRecord.Valid)
}
