// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainIPModel represents the resource state for a Mailgun domain IP association.
type DomainIPModel struct {
	Id     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	IP     types.String `tfsdk:"ip"`
}
