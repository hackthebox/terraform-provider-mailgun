// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package send_alerts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mgerr"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/send_alerts"
)

func testSendAlertsAPIClient(t *testing.T, handler http.HandlerFunc) *send_alerts.SendAlertsAPIClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	return send_alerts.NewSendAlertsAPIClient(mg)
}

func TestGetSendAlert_Found(t *testing.T) {
	client := testSendAlertsAPIClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"my-alert","metric":"delivered_rate","limit":"0.05","dimension":"domain"}`))
	})

	alert, err := client.GetSendAlert(context.Background(), "my-alert")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert.Name != "my-alert" {
		t.Errorf("Name = %q, want %q", alert.Name, "my-alert")
	}
}

// A 404 must surface as a typed error the caller can detect with
// mgerr.IsNotFound, not as a (nil, nil) result that conflates "absent" with
// "success".
func TestGetSendAlert_NotFoundIsTypedError(t *testing.T) {
	client := testSendAlertsAPIClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"alert not found"}`))
	})

	alert, err := client.GetSendAlert(context.Background(), "missing-alert")
	if err == nil {
		t.Fatal("expected a non-nil error for a 404 response")
	}
	if !mgerr.IsNotFound(err) {
		t.Errorf("expected mgerr.IsNotFound(err) = true, got err = %v", err)
	}
	if alert != nil {
		t.Errorf("expected nil alert, got %+v", alert)
	}
}

func TestGetSendAlert_500IsNotTreatedAsNotFound(t *testing.T) {
	client := testSendAlertsAPIClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	_, err := client.GetSendAlert(context.Background(), "my-alert")
	if err == nil {
		t.Fatal("expected an error for a 500 response")
	}
	if mgerr.IsNotFound(err) {
		t.Error("expected mgerr.IsNotFound(err) = false for a 500 response")
	}
}
