// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
)

func int64Attribute(t *testing.T, name string) schema.Int64Attribute {
	t.Helper()
	attr, ok := DomainSendingKeyResourceSchema().Attributes[name].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("attribute %q is missing or not an Int64Attribute", name)
	}
	return attr
}

func requiresReplace(attr schema.Int64Attribute) bool {
	want := int64planmodifier.RequiresReplace().Description(context.Background())
	for _, m := range attr.PlanModifiers {
		if m.Description(context.Background()) == want {
			return true
		}
	}
	return false
}

// deletion_delay only matters at delete time, so changing it must update in place.
// A replacement would rotate the key just to change how long a later delete waits.
func TestSchema_DeletionDelayIsOptionalDefaultsToZeroAndUpdatesInPlace(t *testing.T) {
	attr := int64Attribute(t, "deletion_delay")

	if !attr.Optional {
		t.Error("deletion_delay must be optional")
	}
	if attr.Default == nil {
		t.Fatal("deletion_delay must have a default")
	}
	resp := &defaults.Int64Response{}
	attr.Default.DefaultInt64(context.Background(), defaults.Int64Request{}, resp)
	if got := resp.PlanValue.ValueInt64(); got != 0 {
		t.Errorf("deletion_delay default = %d, want 0", got)
	}
	if requiresReplace(attr) {
		t.Error("deletion_delay must not force replacement")
	}
	if len(attr.Validators) == 0 {
		t.Error("deletion_delay must reject negative values")
	}
}

// Update cannot change a Mailgun key, so a changed expiration must plan a replacement.
func TestSchema_ExpirationRequiresReplace(t *testing.T) {
	if !requiresReplace(int64Attribute(t, "expiration")) {
		t.Error("expiration must force replacement")
	}
}
