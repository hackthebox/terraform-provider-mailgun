// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package test_helpers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// A manually added entry found via the allowlist lookup must be left alone:
// no create call to re-add it, no cleanup call to remove it. If the
// lookup's ctx were lost (e.g. swapped for nil), request construction would
// fail before ever reaching the server, the lookup would report a scan miss
// instead of a match, and SetupIPAllowlistForTests would wrongly create
// (and later delete) an entry the test runner doesn't own.
func TestSetupIPAllowlistForTests_ManuallyAddedEntryIsLeftAlone(t *testing.T) {
	const fakeIP = "203.0.113.9"

	ipSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(fakeIP))
	}))
	t.Cleanup(ipSrv.Close)

	var mu sync.Mutex
	var methods []string

	mailgunSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		methods = append(methods, r.Method)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"addresses":[{"ip_address":"` + fakeIP + `","description":"manually-added"}]}`))
		}
	}))
	t.Cleanup(mailgunSrv.Close)

	origIPURL := getPublicIPURL
	origBaseOverride := testAPIBaseOverride
	getPublicIPURL = ipSrv.URL
	testAPIBaseOverride = mailgunSrv.URL
	t.Cleanup(func() {
		getPublicIPURL = origIPURL
		testAPIBaseOverride = origBaseOverride
	})

	// Registered before SetupIPAllowlistForTests so it runs after that call's
	// own t.Cleanup (LIFO order), capturing any delete it schedules.
	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		if len(methods) != 1 || methods[0] != http.MethodGet {
			t.Errorf("expected only a single GET lookup against the fake server, got %v", methods)
		}
	})

	t.Setenv("MAILGUN_API_KEY", "test-key")

	got := SetupIPAllowlistForTests(t)

	if got != fakeIP {
		t.Fatalf("expected the fixture to return %q, got %q", fakeIP, got)
	}
}
