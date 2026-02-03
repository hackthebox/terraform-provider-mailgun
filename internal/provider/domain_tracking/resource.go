// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

var (
	_ resource.Resource                = &domainTrackingResource{}
	_ resource.ResourceWithConfigure   = &domainTrackingResource{}
	_ resource.ResourceWithImportState = &domainTrackingResource{}
)

// NewDomainTrackingResource creates a new domain tracking resource.
func NewDomainTrackingResource() resource.Resource {
	return &domainTrackingResource{}
}

type domainTrackingResource struct {
	client *mailgun.Client
}

func (r *domainTrackingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_tracking"
}

func (r *domainTrackingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainTrackingResourceSchema()
}

func (r *domainTrackingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainTrackingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainTrackingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	// Update tracking settings
	if err := r.updateTrackingSettings(ctx, domain, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Domain Tracking",
			fmt.Sprintf("Could not configure tracking for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Read back to get computed values
	if err := r.readTrackingSettings(ctx, domain, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Tracking",
			fmt.Sprintf("Could not read tracking settings for domain %s: %s", domain, err.Error()),
		)
		return
	}

	plan.Id = types.StringValue(domain)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainTrackingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainTrackingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	if err := r.readTrackingSettings(ctx, domain, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Tracking",
			fmt.Sprintf("Could not read tracking settings for domain %s: %s", domain, err.Error()),
		)
		return
	}

	state.Id = types.StringValue(domain)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *domainTrackingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainTrackingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	// Update tracking settings
	if err := r.updateTrackingSettings(ctx, domain, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Domain Tracking",
			fmt.Sprintf("Could not update tracking for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Read back to get computed values
	if err := r.readTrackingSettings(ctx, domain, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Tracking",
			fmt.Sprintf("Could not read tracking settings for domain %s: %s", domain, err.Error()),
		)
		return
	}

	plan.Id = types.StringValue(domain)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainTrackingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainTrackingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	// Reset tracking to disabled (best effort)
	resetCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Disable all tracking
	_ = r.client.UpdateClickTracking(resetCtx, domain, "no")
	_ = r.client.UpdateOpenTracking(resetCtx, domain, "no")
	_ = r.client.UpdateUnsubscribeTracking(resetCtx, domain, "no", "", "")

	// Resource removal from state is automatic
}

func (r *domainTrackingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by domain name
	domain := req.ID

	var state DomainTrackingModel
	state.Domain = types.StringValue(domain)

	if err := r.readTrackingSettings(ctx, domain, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Domain Tracking",
			fmt.Sprintf("Could not import tracking settings for domain %s: %s", domain, err.Error()),
		)
		return
	}

	state.Id = types.StringValue(domain)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// updateTrackingSettings updates all tracking settings for a domain
func (r *domainTrackingResource) updateTrackingSettings(ctx context.Context, domain string, model *DomainTrackingModel) error {
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Update click tracking
	clickActive := "no"
	if model.ClickActive.ValueBool() {
		clickActive = "yes"
	}
	if err := r.client.UpdateClickTracking(updateCtx, domain, clickActive); err != nil {
		return fmt.Errorf("failed to update click tracking: %w", err)
	}

	// Update open tracking
	openActive := "no"
	if model.OpenActive.ValueBool() {
		openActive = "yes"
	}
	if err := r.client.UpdateOpenTracking(updateCtx, domain, openActive); err != nil {
		return fmt.Errorf("failed to update open tracking: %w", err)
	}

	// Update unsubscribe tracking
	unsubActive := "no"
	if model.UnsubscribeActive.ValueBool() {
		unsubActive = "yes"
	}
	htmlFooter := ""
	if !model.UnsubscribeHtmlFooter.IsNull() {
		htmlFooter = model.UnsubscribeHtmlFooter.ValueString()
	}
	textFooter := ""
	if !model.UnsubscribeTextFooter.IsNull() {
		textFooter = model.UnsubscribeTextFooter.ValueString()
	}
	if err := r.client.UpdateUnsubscribeTracking(updateCtx, domain, unsubActive, htmlFooter, textFooter); err != nil {
		return fmt.Errorf("failed to update unsubscribe tracking: %w", err)
	}

	return nil
}

// readTrackingSettings reads tracking settings from the API and updates the model
func (r *domainTrackingResource) readTrackingSettings(ctx context.Context, domain string, model *DomainTrackingModel) error {
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tracking, err := r.client.GetDomainTracking(readCtx, domain)
	if err != nil {
		return err
	}

	mapTrackingToModel(&tracking, model)
	return nil
}

// mapTrackingToModel maps Mailgun tracking settings to the Terraform model
func mapTrackingToModel(tracking *mtypes.DomainTracking, model *DomainTrackingModel) {
	model.ClickActive = types.BoolValue(tracking.Click.Active)
	model.OpenActive = types.BoolValue(tracking.Open.Active)
	model.UnsubscribeActive = types.BoolValue(tracking.Unsubscribe.Active)

	if tracking.Unsubscribe.HTMLFooter != "" {
		model.UnsubscribeHtmlFooter = types.StringValue(tracking.Unsubscribe.HTMLFooter)
	} else {
		model.UnsubscribeHtmlFooter = types.StringNull()
	}

	if tracking.Unsubscribe.TextFooter != "" {
		model.UnsubscribeTextFooter = types.StringValue(tracking.Unsubscribe.TextFooter)
	} else {
		model.UnsubscribeTextFooter = types.StringNull()
	}
}
