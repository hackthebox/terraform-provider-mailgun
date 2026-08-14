// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package schema_validators

import (
	"context"
	"fmt"
	"net/mail"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = stringValueValidator{}

type stringValueValidator struct {
	description string
	isValid     func(string) bool
}

func (v stringValueValidator) Description(context.Context) string {
	return v.description
}

func (v stringValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringValueValidator) ValidateString(
	_ context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if v.isValid(value) {
		return
	}

	response.Diagnostics.AddAttributeError(
		request.Path,
		"Invalid Attribute Value",
		fmt.Sprintf("Value %q is invalid: %s.", value, v.description),
	)
}

// IPAddress returns a validator that accepts IPv4 and IPv6 addresses.
func IPAddress() validator.String {
	return stringValueValidator{
		description: "value must be a valid IPv4 or IPv6 address",
		isValid:     isIPAddress,
	}
}

// IPAddressOrCIDR returns a validator that accepts IP addresses and CIDR prefixes.
func IPAddressOrCIDR() validator.String {
	return stringValueValidator{
		description: "value must be a valid IPv4 or IPv6 address or CIDR prefix",
		isValid: func(value string) bool {
			if isIPAddress(value) {
				return true
			}

			prefix, err := netip.ParsePrefix(value)
			return err == nil && prefix.Addr().Zone() == ""
		},
	}
}

// EmailAddress returns a validator that accepts an email address without a display name.
func EmailAddress() validator.String {
	return stringValueValidator{
		description: "value must be a valid email address without a display name",
		isValid: func(value string) bool {
			address, err := mail.ParseAddress(value)
			return err == nil && address.Address == value
		},
	}
}

func isIPAddress(value string) bool {
	address, err := netip.ParseAddr(value)
	return err == nil && address.Zone() == ""
}
