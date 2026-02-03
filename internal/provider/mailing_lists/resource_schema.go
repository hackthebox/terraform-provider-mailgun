// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// MailingListResourceSchema returns the schema for the mailgun_mailing_list resource.
func MailingListResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Mailgun mailing list. Mailing lists allow you to create distribution lists for sending emails to multiple recipients.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "The email address of the mailing list (e.g., developers@lists.example.com). Changing this forces recreation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A display name for the mailing list.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the mailing list.",
				Optional:    true,
			},
			"access_level": schema.StringAttribute{
				Description: "Access level for the mailing list. Valid values: 'readonly' (only admins can post), 'members' (only members can post), 'everyone' (anyone can post). Defaults to 'readonly'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("readonly"),
				Validators: []validator.String{
					stringvalidator.OneOf("readonly", "members", "everyone"),
				},
			},
			"reply_preference": schema.StringAttribute{
				Description: "Where replies should be sent. Valid values: 'list' (replies go to the list), 'sender' (replies go to the original sender). Defaults to 'list'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("list"),
				Validators: []validator.String{
					stringvalidator.OneOf("list", "sender"),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the mailing list (same as address).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the mailing list was created.",
				Computed:    true,
			},
			"members_count": schema.Int64Attribute{
				Description: "The number of members in the mailing list.",
				Computed:    true,
			},
		},
	}
}
