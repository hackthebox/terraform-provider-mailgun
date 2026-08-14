// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mailgun/mailgun-go/v5"
)

func TestPutTrackingSettingSendsMultipart(t *testing.T) {
	var (
		gotContentType string
		gotPath        string
		gotMethod      string
		gotUser        string
		gotPass        string
		gotFields      map[string]string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotUser, gotPass, _ = r.BasicAuth()

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("server could not parse the body as multipart: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotFields = map[string]string{}
		for k, v := range r.MultipartForm.Value {
			gotFields[k] = v[0]
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := putTrackingSetting(context.Background(), server.Client(), server.URL, "key-test", "example.com", "unsubscribe",
		map[string]string{"active": "yes", "html_footer": "<p>x</p>", "text_footer": "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("Content-Type = %q, want multipart/form-data as Mailgun documents", gotContentType)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	if gotPath != "/v3/domains/example.com/tracking/unsubscribe" {
		t.Errorf("path = %s, want the v3 tracking sub-resource", gotPath)
	}
	if gotUser != "api" || gotPass != "key-test" {
		t.Errorf("basic auth = %q/%q, want api/key-test", gotUser, gotPass)
	}
	if gotFields["active"] != "yes" || gotFields["html_footer"] != "<p>x</p>" || gotFields["text_footer"] != "x" {
		t.Errorf("fields = %v, want all three carried verbatim", gotFields)
	}
}

func TestPutTrackingSettingRejectsErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"not allowed"}`))
	}))
	defer server.Close()

	err := putTrackingSetting(context.Background(), server.Client(), server.URL, "key-test", "example.com", "click",
		map[string]string{"active": "yes"})

	if err == nil {
		t.Fatal("expected an error for a non-2xx response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want the status surfaced", err)
	}
}

func TestPutTrackingSettingPreservesEmptyValues(t *testing.T) {
	var gotFields map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("server could not parse the body as multipart: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotFields = map[string]string{}
		for k, v := range r.MultipartForm.Value {
			gotFields[k] = v[0]
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Mailgun treats an omitted param as "keep current", so empties must be sent.
	err := putTrackingSetting(context.Background(), server.Client(), server.URL, "key-test", "example.com", "unsubscribe",
		map[string]string{"active": "no", "html_footer": "", "text_footer": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := gotFields["html_footer"]; !ok {
		t.Error("html_footer must be sent even when empty, or Mailgun keeps the previous footer")
	}
	if _, ok := gotFields["text_footer"]; !ok {
		t.Error("text_footer must be sent even when empty, or Mailgun keeps the previous footer")
	}
}

func TestPutTrackingSettingErrorsOnUnbuildableRequest(t *testing.T) {
	err := putTrackingSetting(context.Background(), http.DefaultClient, "://not a url", "key", "example.com", "click",
		map[string]string{"active": "yes"})

	if err == nil {
		t.Fatal("expected an error for an unbuildable request URL")
	}
	if !strings.Contains(err.Error(), "building click tracking request") {
		t.Errorf("err = %v, want it to name the failing stage", err)
	}
}

func TestPutTrackingSettingErrorsOnTransportFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	err := putTrackingSetting(context.Background(), http.DefaultClient, url, "key", "example.com", "open",
		map[string]string{"active": "yes"})

	if err == nil {
		t.Fatal("expected an error when the server is unreachable")
	}
	if !strings.Contains(err.Error(), "updating open tracking") {
		t.Errorf("err = %v, want it to name the failing stage", err)
	}
}

func TestPutTrackingUsesTheConfiguredClient(t *testing.T) {
	var gotPath, gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		user, pass, _ := r.BasicAuth()
		gotAuth = user + ":" + pass
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mg := mailgun.NewMailgun("key-configured")
	if err := mg.SetAPIBase(server.URL); err != nil {
		t.Fatalf("setting API base: %v", err)
	}

	res := &domainTrackingResource{client: mg}
	if err := res.putTracking(context.Background(), "example.com", "unsubscribe", map[string]string{"active": "no"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v3/domains/example.com/tracking/unsubscribe" {
		t.Errorf("path = %s, want the configured base to be used verbatim", gotPath)
	}
	if gotAuth != "api:key-configured" {
		t.Errorf("auth = %s, want the client's API key", gotAuth)
	}
}
