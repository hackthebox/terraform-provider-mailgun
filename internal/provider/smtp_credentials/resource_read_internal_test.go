// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

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

func TestFindCredential_ReturnsFoundTrueWhenPresent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":1,"items":[{"login":"user@example.com"}]}`))
	})

	r := &SmtpCredentialResource{client: client}
	cred, found, err := r.findCredential(context.Background(), "example.com", "user")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected the credential to be found")
	}
	if cred.Login != "user@example.com" {
		t.Errorf("cred.Login = %q, want %q", cred.Login, "user@example.com")
	}
}

func TestFindCredential_ReturnsFoundFalseWhenAbsent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
	})

	r := &SmtpCredentialResource{client: client}
	_, found, err := r.findCredential(context.Background(), "example.com", "gone-user")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected the credential not to be found")
	}
}

func TestFindCredential_ReturnsErrorOnListFailure(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	r := &SmtpCredentialResource{client: client}
	_, found, err := r.findCredential(context.Background(), "example.com", "user")

	if err == nil {
		t.Fatal("expected an error when the listing itself fails")
	}
	if found {
		t.Error("found must be false when err is non-nil")
	}
	// The wire status must survive the wrap (%w, not %v), or callers using
	// errors.As/Is (and mgerr.IsNotFound, transitively) lose the cause.
	if got := mailgun.GetStatusFromErr(err); got != http.StatusInternalServerError {
		t.Errorf("GetStatusFromErr(err) = %d, want %d (the wire status must survive the wrap)", got, http.StatusInternalServerError)
	}
}

func readCredential(t *testing.T, client *mailgun.Client, domain, login string) *resource.ReadResponse {
	t.Helper()

	resourceSchema := SmtpCredentialResourceSchema()
	raw := credentialObject(t, map[string]tftypes.Value{
		"domain": tftypes.NewValue(tftypes.String, domain),
		"login":  tftypes.NewValue(tftypes.String, login),
	})
	state := tfsdk.State{Raw: raw, Schema: resourceSchema}

	resp := &resource.ReadResponse{State: state}
	(&SmtpCredentialResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func TestRead_RemovesResourceWhenCredentialAbsent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
	})

	resp := readCredential(t, client, "example.com", "gone-user")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state when the credential is absent")
	}
}

// A listing failure (500, rate limit, transport error) is not the same as
// the credential being gone: it must produce a diagnostic and leave state
// intact rather than silently dropping a live resource.
func TestRead_ListFailureProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readCredential(t, client, "example.com", "user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the credential listing fails")
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

	resp := readCredential(t, mg, "example.com", "user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a transport failure")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a transport failure must not remove the resource from state")
	}
}
