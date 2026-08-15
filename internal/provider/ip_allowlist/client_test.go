// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/ip_allowlist"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mgerr"
)

func testIPAllowlistClient(t *testing.T, handler http.HandlerFunc) *ip_allowlist.IPAllowlistClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	return ip_allowlist.NewIPAllowlistClient(mg)
}

func TestGetIPAllowlistEntry_FoundInListing(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"addresses":[{"ip_address":"203.0.113.1","description":"present"}]}`))
	})

	entry, found, err := client.GetIPAllowlistEntry(context.Background(), "203.0.113.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected found = true")
	}
	if entry.Description != "present" {
		t.Errorf("Description = %q, want %q", entry.Description, "present")
	}
}

// A 200 listing that simply does not contain the address is a scan miss, not
// a wire-level error: found is false and err is nil.
func TestGetIPAllowlistEntry_ScanMiss(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"addresses":[]}`))
	})

	entry, found, err := client.GetIPAllowlistEntry(context.Background(), "203.0.113.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected found = false")
	}
	if entry != nil {
		t.Errorf("expected nil entry, got %+v", entry)
	}
}

// A genuine transport/API-level failure on the underlying List call must
// surface as an error, never as a silent scan miss, even if its body happens
// to mention "not found" or "404".
func TestGetIPAllowlistEntry_ListingErrorPropagates(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"entry not found (404)"}`))
	})

	entry, found, err := client.GetIPAllowlistEntry(context.Background(), "203.0.113.1")
	if err == nil {
		t.Fatal("expected an error for a 500 response")
	}
	if found {
		t.Error("expected found = false alongside the error")
	}
	if entry != nil {
		t.Errorf("expected nil entry, got %+v", entry)
	}
}

func TestDeleteIPAllowlistEntry_GenuineNotFound(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"entry not found"}`))
	})

	err := client.DeleteIPAllowlistEntry(context.Background(), "203.0.113.1")
	if !mgerr.IsNotFound(err) {
		t.Fatalf("expected mgerr.IsNotFound(err) = true, got err = %v", err)
	}
}

// A 500 whose body happens to contain the literal "404" must not resolve as
// IsNotFound: only the real HTTP status counts.
func TestDeleteIPAllowlistEntry_500MentioningLiteral404(t *testing.T) {
	client := testIPAllowlistClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream returned 404 from internal service"}`))
	})

	err := client.DeleteIPAllowlistEntry(context.Background(), "203.0.113.1")
	if err == nil {
		t.Fatal("expected an error for a 500 response")
	}
	if mgerr.IsNotFound(err) {
		t.Error("expected mgerr.IsNotFound(err) = false for a 500 response")
	}
}
