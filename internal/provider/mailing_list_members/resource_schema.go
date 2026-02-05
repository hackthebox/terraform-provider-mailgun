// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MailingListMemberResourceSchema returns the schema for the mailgun_mailing_list_member resource.
func MailingListMemberResourceSchema() schema.Schema {
	return schema.Schema{
		Version:     0,
		Description: "Manages a member of a Mailgun mailing list. Members are email addresses that receive messages sent to the mailing list.",
		Attributes: map[string]schema.Attribute{
			"list_address": schema.StringAttribute{
				Description: "The email address of the mailing list. Changing this forces recreation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_address": schema.StringAttribute{
				Description: "The email address of the member. Changing this forces recreation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the member.",
				Optional:    true,
			},
			"subscribed": schema.BoolAttribute{
				Description: "Whether the member is subscribed to receive emails. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"vars": schema.MapAttribute{
				Description: "A map of custom variables for the member. These can be used in email templates.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the member (list_address/member_address).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
