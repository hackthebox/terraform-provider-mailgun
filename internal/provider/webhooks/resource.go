// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

// NewWebhookResource creates a new webhook resource.
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	client *mailgun.Client
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WebhookResourceSchema()
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert URLs list to []string
	var urls []string
	diags = plan.Urls.ElementsAs(ctx, &urls, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	webhookType := plan.WebhookType.ValueString()

	// Create the webhook
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.CreateWebhook(createCtx, domain, webhookType, urls)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mailgun Webhook",
			fmt.Sprintf("Could not create webhook %s for domain %s: %s", webhookType, domain, err.Error()),
		)
		return
	}

	// Set computed fields
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, webhookType))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	webhookType := state.WebhookType.ValueString()

	// Get webhook from API
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	urls, err := r.client.GetWebhook(readCtx, domain, webhookType)
	if err != nil {
		// Check if webhook doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Mailgun Webhook",
			fmt.Sprintf("Could not read webhook %s for domain %s: %s", webhookType, domain, err.Error()),
		)
		return
	}

	// Update state with fetched data
	urlsList, urlsDiags := types.ListValueFrom(ctx, types.StringType, urls)
	resp.Diagnostics.Append(urlsDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Urls = urlsList
	state.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, webhookType))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert URLs list to []string
	var urls []string
	diags = plan.Urls.ElementsAs(ctx, &urls, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	webhookType := plan.WebhookType.ValueString()

	// Update the webhook
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.UpdateWebhook(updateCtx, domain, webhookType, urls)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mailgun Webhook",
			fmt.Sprintf("Could not update webhook %s for domain %s: %s", webhookType, domain, err.Error()),
		)
		return
	}

	// Set computed fields
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, webhookType))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	webhookType := state.WebhookType.ValueString()

	// Delete the webhook
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.client.DeleteWebhook(deleteCtx, domain, webhookType)
	if err != nil {
		// Ignore not found errors during delete
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error Deleting Mailgun Webhook",
				fmt.Sprintf("Could not delete webhook %s for domain %s: %s", webhookType, domain, err.Error()),
			)
			return
		}
	}
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: domain/webhook_type
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'domain/webhook_type', got: %s", req.ID),
		)
		return
	}

	domain := parts[0]
	webhookType := parts[1]

	// Validate webhook type
	validTypes := map[string]bool{
		"accepted":       true,
		"delivered":      true,
		"permanent_fail": true,
		"temporary_fail": true,
		"opened":         true,
		"clicked":        true,
		"unsubscribed":   true,
		"complained":     true,
	}
	if !validTypes[webhookType] {
		resp.Diagnostics.AddError(
			"Invalid Webhook Type",
			fmt.Sprintf("Invalid webhook type '%s'. Valid types: accepted, delivered, permanent_fail, temporary_fail, opened, clicked, unsubscribed, complained", webhookType),
		)
		return
	}

	// Get webhook from API
	importCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	urls, err := r.client.GetWebhook(importCtx, domain, webhookType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Mailgun Webhook",
			fmt.Sprintf("Could not import webhook %s for domain %s: %s", webhookType, domain, err.Error()),
		)
		return
	}

	// Build state
	urlsList, urlsDiags := types.ListValueFrom(ctx, types.StringType, urls)
	resp.Diagnostics.Append(urlsDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := WebhookModel{
		Domain:      types.StringValue(domain),
		WebhookType: types.StringValue(webhookType),
		Urls:        urlsList,
		Id:          types.StringValue(req.ID),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
