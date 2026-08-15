// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domains

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

// domainStateObject builds a resource-shaped tftypes.Value, defaulting every
// attribute to null and applying the given overrides.
func domainStateObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := DomainResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
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

func readDomain(t *testing.T, client *mailgun.Client, name string) *resource.ReadResponse {
	t.Helper()

	resourceSchema := DomainResourceSchema()
	state := tfsdk.State{
		Raw:    domainStateObject(t, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, name)}),
		Schema: resourceSchema,
	}

	resp := &resource.ReadResponse{State: state}
	(&DomainResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func TestRead_RemovesResourceOnGenuine404(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Domain not found"}`))
	})

	resp := readDomain(t, client, "gone.example.com")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state on a genuine 404")
	}
}

func TestRead_500ProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readDomain(t, client, "test.example.com")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a transport/server error must not remove the resource from state")
	}
}

func TestRead_RateLimitedProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"Too Many Requests"}`))
	})

	resp := readDomain(t, client, "test.example.com")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 429 response")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a rate-limited error must not remove the resource from state")
	}
}

func TestRead_TransportErrorProducesErrorAndLeavesStateIntact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	srv.Close() // connections to this address now fail outright

	resp := readDomain(t, mg, "test.example.com")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a transport failure")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a transport failure must not remove the resource from state")
	}
}
