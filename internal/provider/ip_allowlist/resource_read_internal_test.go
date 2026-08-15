// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

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

func testIPAllowlistClient(t *testing.T, handler http.HandlerFunc) *IPAllowlistClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	return NewIPAllowlistClient(mg)
}

// ipAllowlistStateObject builds a resource-shaped tftypes.Value, defaulting
// every attribute to null and applying the given overrides.
func ipAllowlistStateObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := IPAllowlistResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
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

func readIPAllowlist(t *testing.T, client *IPAllowlistClient, address string) *resource.ReadResponse {
	t.Helper()

	resourceSchema := IPAllowlistResourceSchema()
	state := tfsdk.State{
		Raw: ipAllowlistStateObject(t, map[string]tftypes.Value{
			"address":     tftypes.NewValue(tftypes.String, address),
			"description": tftypes.NewValue(tftypes.String, "test"),
			"id":          tftypes.NewValue(tftypes.String, address),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.ReadResponse{State: state}
	(&ipAllowlistResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func deleteIPAllowlist(t *testing.T, client *IPAllowlistClient, address string) *resource.DeleteResponse {
	t.Helper()

	resourceSchema := IPAllowlistResourceSchema()
	state := tfsdk.State{
		Raw: ipAllowlistStateObject(t, map[string]tftypes.Value{
			"address":     tftypes.NewValue(tftypes.String, address),
			"description": tftypes.NewValue(tftypes.String, "test"),
			"id":          tftypes.NewValue(tftypes.String, address),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.DeleteResponse{State: state}
	(&ipAllowlistResource{client: client}).Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	return resp
}

// A 500 whose body happens to mention "not found (404)" must not be mistaken
// for a scan miss: the strings.Contains this replaces would remove a live
// resource here.
func TestRead_500MentioningNotFoundIsNotTreatedAsMissing(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"entry not found (404)"}`))
	})

	resp := readIPAllowlist(t, client, "203.0.113.1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response, even one mentioning 'not found (404)'")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a 500 must not remove the resource from state just because its body mentions 'not found'")
	}
}

// A 200 listing that simply omits the address is a genuine scan miss: the
// resource should be removed from state with no diagnostics.
func TestRead_ScanMissRemovesResource(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"addresses":[]}`))
	})

	resp := readIPAllowlist(t, client, "203.0.113.1")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state on a scan miss")
	}
}

func TestDelete_SuccessProducesNoDiagnostics(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	resp := deleteIPAllowlist(t, client, "203.0.113.1")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for a successful delete: %v", resp.Diagnostics)
	}
}

func TestDelete_GenuineNotFoundIsIgnored(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"entry not found"}`))
	})

	resp := deleteIPAllowlist(t, client, "203.0.113.1")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for a genuine 404 during delete: %v", resp.Diagnostics)
	}
}

// A 500 whose body happens to contain the literal "404" must raise a
// diagnostic instead of being silently ignored as "already deleted".
func TestDelete_500MentioningLiteral404RaisesDiagnostic(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream returned 404 from internal service"}`))
	})

	resp := deleteIPAllowlist(t, client, "203.0.113.1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response mentioning the literal '404'")
	}
}

func importIPAllowlist(t *testing.T, client *IPAllowlistClient, address string) *resource.ImportStateResponse {
	t.Helper()

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: IPAllowlistResourceSchema()},
	}
	(&ipAllowlistResource{client: client}).ImportState(context.Background(), resource.ImportStateRequest{ID: address}, resp)
	return resp
}

func TestImportState_Found(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"addresses":[{"ip_address":"203.0.113.1","description":"present"}]}`))
	})

	resp := importIPAllowlist(t, client, "203.0.113.1")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Error("expected state to be populated on a successful import")
	}
}

func TestImportState_ListingErrorRaisesDiagnostic(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := importIPAllowlist(t, client, "203.0.113.1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response")
	}
}

func TestImportState_ScanMissRaisesDiagnostic(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"addresses":[]}`))
	})

	resp := importIPAllowlist(t, client, "203.0.113.1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the address is not in the allowlist")
	}
}
