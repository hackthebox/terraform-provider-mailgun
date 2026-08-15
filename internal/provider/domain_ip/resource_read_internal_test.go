// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_ip

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

// domainIPStateObject builds a resource-shaped tftypes.Value, defaulting
// every attribute to null and applying the given overrides.
func domainIPStateObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := DomainIPResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
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

func readDomainIP(t *testing.T, client *mailgun.Client, domain, ip string) *resource.ReadResponse {
	t.Helper()

	resourceSchema := DomainIPResourceSchema()
	state := tfsdk.State{
		Raw: domainIPStateObject(t, map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, domain),
			"ip":     tftypes.NewValue(tftypes.String, ip),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.ReadResponse{State: state}
	(&domainIPResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

// A genuine 404 from the domain-scoped IP listing means the parent domain
// was deleted out of band, not that the listing merely failed; Read must
// recover the same way it does for an IP that is simply no longer assigned.
func TestRead_RemovesResourceOnGenuine404(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Domain not found"}`))
	})

	resp := readDomainIP(t, client, "gone.example.com", "10.0.0.1")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state on a genuine 404")
	}
}

// A listing failure (500, rate limit, transport error) is not the same as
// the domain being gone: it must produce a diagnostic and leave state
// intact rather than silently dropping a live resource.
func TestRead_ListFailureProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readDomainIP(t, client, "example.com", "10.0.0.1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the IP listing fails")
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

	resp := readDomainIP(t, mg, "example.com", "10.0.0.1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a transport failure")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a transport failure must not remove the resource from state")
	}
}
