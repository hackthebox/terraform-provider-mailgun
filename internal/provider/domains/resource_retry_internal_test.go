// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domains

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func okDomainResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"domain":{"name":"test.example.com","state":"active"}}`))
}

// Mailgun is eventually consistent: a GET can 404 right after a successful
// write. getDomainWithRetry must absorb that transient 404.
func TestGetDomainWithRetry_RetriesTransient404(t *testing.T) {
	var calls int
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Domain not found"}`))
			return
		}
		okDomainResponse(w)
	})

	resp, err := getDomainWithRetry(t.Context(), client, "test.example.com", 5, time.Millisecond)
	if err != nil {
		t.Fatalf("expected success after transient 404s, got error: %v", err)
	}
	if resp.Domain.Name != "test.example.com" {
		t.Errorf("expected domain name to be set, got %q", resp.Domain.Name)
	}
	if calls != 3 {
		t.Errorf("expected 3 attempts, got %d", calls)
	}
}

func TestGetDomainWithRetry_Persistent404Errors(t *testing.T) {
	var calls int
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Domain not found"}`))
	})

	_, err := getDomainWithRetry(t.Context(), client, "test.example.com", 3, time.Millisecond)
	if err == nil {
		t.Fatal("expected an error for persistent 404")
	}
	if got := mailgun.GetStatusFromErr(err); got != http.StatusNotFound {
		t.Errorf("expected 404 status in error, got %d", got)
	}
	if calls != 3 {
		t.Errorf("expected 3 attempts, got %d", calls)
	}
}

func TestGetDomainWithRetry_SuccessFirstTry(t *testing.T) {
	var calls int
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		okDomainResponse(w)
	})

	if _, err := getDomainWithRetry(t.Context(), client, "test.example.com", 5, time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 attempt, got %d", calls)
	}
}

func TestSleepWithContext_ReturnsAfterDelay(t *testing.T) {
	if err := sleepWithContext(t.Context(), time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSleepWithContext_ReturnsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := sleepWithContext(ctx, time.Hour); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

// Non-404 errors are not eventual-consistency artifacts and must not be retried.
func TestGetDomainWithRetry_Non404NotRetried(t *testing.T) {
	var calls int
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	_, err := getDomainWithRetry(t.Context(), client, "test.example.com", 5, time.Millisecond)
	if err == nil {
		t.Fatal("expected an error for 500")
	}
	if calls != 1 {
		t.Errorf("expected no retry on 500, got %d attempts", calls)
	}
}
