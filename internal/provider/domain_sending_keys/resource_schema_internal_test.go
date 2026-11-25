// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func int64Attribute(t *testing.T, name string) schema.Int64Attribute {
	t.Helper()
	attr, ok := DomainSendingKeyResourceSchema().Attributes[name].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("attribute %q is missing or not an Int64Attribute", name)
	}
	return attr
}

func requiresReplace(t *testing.T, attr schema.Attribute) bool {
	t.Helper()
	ctx := context.Background()
	switch a := attr.(type) {
	case schema.Int64Attribute:
		want := int64planmodifier.RequiresReplace().Description(ctx)
		for _, m := range a.PlanModifiers {
			if m.Description(ctx) == want {
				return true
			}
		}
	case schema.StringAttribute:
		want := stringplanmodifier.RequiresReplace().Description(ctx)
		for _, m := range a.PlanModifiers {
			if m.Description(ctx) == want {
				return true
			}
		}
	default:
		t.Fatalf("unexpected attribute type %T", attr)
	}
	return false
}

func deletionDelayAccepts(attr schema.Int64Attribute, value int64) bool {
	req := validator.Int64Request{ConfigValue: types.Int64Value(value)}
	resp := &validator.Int64Response{}
	for _, v := range attr.Validators {
		v.ValidateInt64(context.Background(), req, resp)
	}
	return !resp.Diagnostics.HasError()
}

// deletion_delay has no default: an unset value stays null, so upgrading the
// provider plans no diff on existing keys. Changing it must update in place, or
// a new delay would rotate the key.
func TestSchema_DeletionDelayIsOptionalWithoutDefaultAndUpdatesInPlace(t *testing.T) {
	attr := int64Attribute(t, "deletion_delay")

	if !attr.Optional {
		t.Error("deletion_delay must be optional")
	}
	if attr.Computed || attr.Default != nil {
		t.Error("deletion_delay must not be computed or have a default")
	}
	if requiresReplace(t, attr) {
		t.Error("deletion_delay must not force replacement")
	}
}

func TestSchema_DeletionDelayBounds(t *testing.T) {
	attr := int64Attribute(t, "deletion_delay")

	for value, want := range map[int64]bool{-1: false, 0: true, 3600: true, 3601: false} {
		if got := deletionDelayAccepts(attr, value); got != want {
			t.Errorf("deletion_delay = %d accepted: %t, want %t", value, got, want)
		}
	}
}

// Update writes the plan to state without calling Mailgun, which is only safe
// while every attribute Mailgun stores forces replacement.
func TestSchema_EveryConfigurableAttributeButDeletionDelayRequiresReplace(t *testing.T) {
	for name, attr := range DomainSendingKeyResourceSchema().Attributes {
		if name == "deletion_delay" || (!attr.IsRequired() && !attr.IsOptional()) {
			continue
		}
		if !requiresReplace(t, attr) {
			t.Errorf("%s must force replacement", name)
		}
	}
}
