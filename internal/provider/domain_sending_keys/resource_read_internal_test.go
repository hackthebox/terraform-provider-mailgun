// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func testClient(t *testing.T, handler http.HandlerFunc) *mailgun.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	return mg
}

func TestFindKey_ReturnsFoundTrueWhenPresent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"id":"key-1","kind":"domain","domain_name":"example.com"}]}`))
	})

	r := &domainSendingKeyResource{client: client}
	key, found, err := r.findKey(context.Background(), "key-1", "example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected the key to be found")
	}
	if key.ID != "key-1" {
		t.Errorf("key.ID = %q, want %q", key.ID, "key-1")
	}
}

func TestFindKey_ReturnsFoundFalseWhenAbsent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	r := &domainSendingKeyResource{client: client}
	_, found, err := r.findKey(context.Background(), "gone-key", "example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected the key not to be found")
	}
}

func TestFindKey_ReturnsErrorOnListFailure(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	r := &domainSendingKeyResource{client: client}
	_, found, err := r.findKey(context.Background(), "key-1", "example.com")

	if err == nil {
		t.Fatal("expected an error when the listing itself fails")
	}
	if found {
		t.Error("found must be false when err is non-nil")
	}
}

// readStateObject builds a resource-shaped tftypes.Value, defaulting every
// attribute to null and applying the given overrides.
func readStateObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := DomainSendingKeyResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("schema terraform type is not an object")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		if override, found := overrides[name]; found {
			attrs[name] = override
			continue
		}
		attrs[name] = tftypes.NewValue(attrType, nil)
	}

	return tftypes.NewValue(objType, attrs)
}

func readSendingKey(t *testing.T, client *mailgun.Client, id, domain string) *resource.ReadResponse {
	t.Helper()

	resourceSchema := DomainSendingKeyResourceSchema()
	state := tfsdk.State{
		Raw: readStateObject(t, map[string]tftypes.Value{
			"id":     tftypes.NewValue(tftypes.String, id),
			"domain": tftypes.NewValue(tftypes.String, domain),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.ReadResponse{State: state}
	(&domainSendingKeyResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func TestRead_RemovesResourceWhenKeyAbsent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	resp := readSendingKey(t, client, "gone-key", "example.com")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state when the key is absent")
	}
}

// A listing failure (500, rate limit, transport error) is not the same as the
// key being gone: it must produce a diagnostic and leave state intact rather
// than silently dropping a live resource.
func TestRead_ListFailureProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readSendingKey(t, client, "key-1", "example.com")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the API key listing fails")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a listing failure must not remove the resource from state")
	}
}

func TestRead_TransportErrorProducesErrorAndLeavesStateIntact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	srv.Close() // connections to this address now fail outright

	resp := readSendingKey(t, mg, "key-1", "example.com")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a transport failure")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a transport failure must not remove the resource from state")
	}
}
