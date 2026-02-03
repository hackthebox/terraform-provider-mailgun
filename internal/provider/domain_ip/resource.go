// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip

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
	_ resource.Resource                = &domainIPResource{}
	_ resource.ResourceWithConfigure   = &domainIPResource{}
	_ resource.ResourceWithImportState = &domainIPResource{}
)

// NewDomainIPResource creates a new domain IP resource.
func NewDomainIPResource() resource.Resource {
	return &domainIPResource{}
}

type domainIPResource struct {
	client *mailgun.Client
}

func (r *domainIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_ip"
}

func (r *domainIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainIPResourceSchema()
}

func (r *domainIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainIPModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	ip := plan.IP.ValueString()

	// Create context with timeout
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Add the IP to the domain
	if err := r.client.AddDomainIP(createCtx, domain, ip); err != nil {
		resp.Diagnostics.AddError(
			"Error Adding IP to Domain",
			fmt.Sprintf("Could not add IP %s to domain %s: %s", ip, domain, err.Error()),
		)
		return
	}

	// Set the ID and state
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", domain, ip))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainIPModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	ip := state.IP.ValueString()

	// Create context with timeout
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List domain IPs and check if our IP exists
	ips, err := r.client.ListDomainIPs(readCtx, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain IPs",
			fmt.Sprintf("Could not read IPs for domain %s: %s", domain, err.Error()),
		)
		return
	}

	// Check if the IP is associated with the domain
	found := false
	for _, domainIP := range ips {
		if domainIP.IP == ip {
			found = true
			break
		}
	}

	if !found {
		// IP no longer associated with domain
		resp.State.RemoveResource(ctx)
		return
	}

	// State remains unchanged
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *domainIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No updates possible - both domain and ip require replacement
	var plan DomainIPModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainIPModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	ip := state.IP.ValueString()

	// Create context with timeout
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.client.DeleteDomainIP(deleteCtx, domain, ip); err != nil {
		resp.Diagnostics.AddError(
			"Error Removing IP from Domain",
			fmt.Sprintf("Could not remove IP %s from domain %s: %s", ip, domain, err.Error()),
		)
		return
	}
}

func (r *domainIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: domain/ip
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format: domain/ip",
		)
		return
	}

	domain := parts[0]
	ip := parts[1]

	// Verify the IP is associated with the domain
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ips, err := r.client.ListDomainIPs(readCtx, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Domain IP",
			fmt.Sprintf("Could not read IPs for domain %s: %s", domain, err.Error()),
		)
		return
	}

	found := false
	for _, domainIP := range ips {
		if domainIP.IP == ip {
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Domain IP Not Found",
			fmt.Sprintf("IP %s is not associated with domain %s", ip, domain),
		)
		return
	}

	state := DomainIPModel{
		Id:     types.StringValue(fmt.Sprintf("%s/%s", domain, ip)),
		Domain: types.StringValue(domain),
		IP:     types.StringValue(ip),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
