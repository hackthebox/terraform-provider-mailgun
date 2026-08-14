// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package schema_validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_dkim_key"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_ip"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/ip_allowlist"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mailing_list_members"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/schema_validators"
)

func TestStringValidators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		validator validator.String
		value     types.String
		wantError bool
	}{
		{
			name:      "IPv4 address",
			validator: schema_validators.IPAddress(),
			value:     types.StringValue("192.0.2.1"),
		},
		{
			name:      "IPv6 address",
			validator: schema_validators.IPAddress(),
			value:     types.StringValue("2001:db8::1"),
		},
		{
			name:      "CIDR is not a bare address",
			validator: schema_validators.IPAddress(),
			value:     types.StringValue("192.0.2.0/24"),
			wantError: true,
		},
		{
			name:      "invalid IP address",
			validator: schema_validators.IPAddress(),
			value:     types.StringValue("192.0.2.999"),
			wantError: true,
		},
		{
			name:      "IPv4 CIDR",
			validator: schema_validators.IPAddressOrCIDR(),
			value:     types.StringValue("192.0.2.0/24"),
		},
		{
			name:      "bare IP is accepted by IP-or-CIDR validator",
			validator: schema_validators.IPAddressOrCIDR(),
			value:     types.StringValue("192.0.2.1"),
		},
		{
			name:      "IPv6 CIDR",
			validator: schema_validators.IPAddressOrCIDR(),
			value:     types.StringValue("2001:db8::/32"),
		},
		{
			name:      "invalid CIDR prefix length",
			validator: schema_validators.IPAddressOrCIDR(),
			value:     types.StringValue("192.0.2.0/33"),
			wantError: true,
		},
		{
			name:      "hostname is not an IP address or CIDR",
			validator: schema_validators.IPAddressOrCIDR(),
			value:     types.StringValue("example.com"),
			wantError: true,
		},
		{
			name:      "email address",
			validator: schema_validators.EmailAddress(),
			value:     types.StringValue("member+tag@example.com"),
		},
		{
			name:      "display name is not a plain email address",
			validator: schema_validators.EmailAddress(),
			value:     types.StringValue("Member <member@example.com>"),
			wantError: true,
		},
		{
			name:      "invalid email address",
			validator: schema_validators.EmailAddress(),
			value:     types.StringValue("member.example.com"),
			wantError: true,
		},
		{
			name:      "null values are deferred",
			validator: schema_validators.IPAddress(),
			value:     types.StringNull(),
		},
		{
			name:      "unknown values are deferred",
			validator: schema_validators.EmailAddress(),
			value:     types.StringUnknown(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var response validator.StringResponse
			test.validator.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("value"),
				ConfigValue: test.value,
			}, &response)

			if got := response.Diagnostics.HasError(); got != test.wantError {
				t.Fatalf("HasError() = %t, want %t; diagnostics: %v", got, test.wantError, response.Diagnostics)
			}
		})
	}
}

func TestResourceSchemasWireValidators(t *testing.T) {
	t.Parallel()

	stringTests := []struct {
		name       string
		attribute  rschema.Attribute
		validValue string
		badValue   string
	}{
		{
			name:       "IP allowlist address",
			attribute:  ip_allowlist.IPAllowlistResourceSchema().Attributes["address"],
			validValue: "192.0.2.0/24",
			badValue:   "192.0.2.999",
		},
		{
			name:       "domain IP address",
			attribute:  domain_ip.DomainIPResourceSchema().Attributes["ip"],
			validValue: "2001:db8::1",
			badValue:   "example.com",
		},
		{
			name:       "mailing list address",
			attribute:  mailing_list_members.MailingListMemberResourceSchema().Attributes["list_address"],
			validValue: "list@example.com",
			badValue:   "list.example.com",
		},
		{
			name:       "mailing list member address",
			attribute:  mailing_list_members.MailingListMemberResourceSchema().Attributes["member_address"],
			validValue: "member@example.com",
			badValue:   "member.example.com",
		},
	}

	for _, test := range stringTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			attribute, ok := test.attribute.(rschema.StringAttribute)
			if !ok {
				t.Fatalf("attribute type = %T, want schema.StringAttribute", test.attribute)
			}
			if len(attribute.Validators) == 0 {
				t.Fatal("attribute has no validators")
			}

			assertStringValidation(t, attribute.Validators, test.validValue, false)
			assertStringValidation(t, attribute.Validators, test.badValue, true)
		})
	}

	t.Run("DKIM key bits", func(t *testing.T) {
		t.Parallel()

		rawAttribute := domain_dkim_key.DomainDkimKeyResourceSchema().Attributes["bits"]
		attribute, ok := rawAttribute.(rschema.Int64Attribute)
		if !ok {
			t.Fatalf("attribute type = %T, want schema.Int64Attribute", rawAttribute)
		}
		if len(attribute.Validators) == 0 {
			t.Fatal("attribute has no validators")
		}

		assertInt64Validation(t, attribute.Validators, 1024, false)
		assertInt64Validation(t, attribute.Validators, 2048, false)
		assertInt64Validation(t, attribute.Validators, 4096, true)
	})
}

func assertStringValidation(
	t *testing.T,
	validators []validator.String,
	value string,
	wantError bool,
) {
	t.Helper()

	var response validator.StringResponse
	request := validator.StringRequest{
		Path:        path.Root("value"),
		ConfigValue: types.StringValue(value),
	}
	for _, attributeValidator := range validators {
		attributeValidator.ValidateString(context.Background(), request, &response)
	}

	if got := response.Diagnostics.HasError(); got != wantError {
		t.Fatalf("validating %q: HasError() = %t, want %t; diagnostics: %v", value, got, wantError, response.Diagnostics)
	}
}

func assertInt64Validation(
	t *testing.T,
	validators []validator.Int64,
	value int64,
	wantError bool,
) {
	t.Helper()

	var response validator.Int64Response
	request := validator.Int64Request{
		Path:        path.Root("value"),
		ConfigValue: types.Int64Value(value),
	}
	for _, attributeValidator := range validators {
		attributeValidator.ValidateInt64(context.Background(), request, &response)
	}

	if got := response.Diagnostics.HasError(); got != wantError {
		t.Fatalf("validating %d: HasError() = %t, want %t; diagnostics: %v", value, got, wantError, response.Diagnostics)
	}
}
