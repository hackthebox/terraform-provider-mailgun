// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

// lookupCreatedCredential must surface a genuine listing failure as its own
// error, not collapse it into the "not found yet" sentinel findEventually
// also retries on: the two must stay distinguishable so a caller inspecting
// the error (e.g. via mailgun.GetStatusFromErr) sees the real cause.
func TestLookupCreatedCredential_ReturnsUnderlyingErrorOnListingFailure(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	r := &SmtpCredentialResource{client: client}
	_, err := r.lookupCreatedCredential(context.Background(), "example.com", "user")

	if err == nil {
		t.Fatal("expected an error when the listing itself fails")
	}
	if errors.Is(err, errCredentialNotFound) {
		t.Error("a listing failure must not be reported as the not-found sentinel")
	}
	if got := mailgun.GetStatusFromErr(err); got != http.StatusInternalServerError {
		t.Errorf("GetStatusFromErr(err) = %d, want %d (the underlying error must survive)", got, http.StatusInternalServerError)
	}
}

func TestLookupCreatedCredential_ReturnsSentinelWhenNotYetVisible(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
	})

	r := &SmtpCredentialResource{client: client}
	_, err := r.lookupCreatedCredential(context.Background(), "example.com", "user")

	if !errors.Is(err, errCredentialNotFound) {
		t.Errorf("err = %v, want errCredentialNotFound", err)
	}
}

func TestLookupCreatedCredential_ReturnsCredentialWhenFound(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(createdCredentialListResponse())
	})

	r := &SmtpCredentialResource{client: client}
	cred, err := r.lookupCreatedCredential(context.Background(), "example.com", "user")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.Login != "user@example.com" {
		t.Errorf("cred.Login = %q, want %q", cred.Login, "user@example.com")
	}
}

func createCredential(t *testing.T, client *mailgun.Client, planOverrides, configOverrides map[string]tftypes.Value) *resource.CreateResponse {
	t.Helper()

	resourceSchema := SmtpCredentialResourceSchema()
	plan := credentialObject(t, planOverrides)
	config := credentialObject(t, configOverrides)

	resp := &resource.CreateResponse{State: tfsdk.State{Raw: plan, Schema: resourceSchema}}
	(&SmtpCredentialResource{client: client}).Create(context.Background(), resource.CreateRequest{
		Plan:   tfsdk.Plan{Raw: plan, Schema: resourceSchema},
		Config: tfsdk.Config{Raw: config, Schema: resourceSchema},
	}, resp)
	return resp
}

func createdCredentialListResponse() []byte {
	return []byte(`{"total_count":1,"items":[{"login":"user@example.com","created_at":"Mon, 02 Jan 2006 15:04:05 UTC"}]}`)
}

func TestCreate_Success_LegacyPassword(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(createdCredentialListResponse())
		}
	})

	resp := createCredential(t, client,
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "legacy-secret"),
		},
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "legacy-secret"),
		},
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestCreate_Success_WriteOnlyPassword(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(createdCredentialListResponse())
		}
	})

	resp := createCredential(t, client,
		map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, "example.com"),
			"login":  tftypes.NewValue(tftypes.String, "user"),
		},
		map[string]tftypes.Value{
			"domain":              tftypes.NewValue(tftypes.String, "example.com"),
			"login":               tftypes.NewValue(tftypes.String, "user"),
			"password_wo":         tftypes.NewValue(tftypes.String, "wo-secret"),
			"password_wo_version": tftypes.NewValue(tftypes.Number, int64(1)),
		},
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestCreate_MissingDomain(t *testing.T) {
	resp := createCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}),
		map[string]tftypes.Value{
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
		map[string]tftypes.Value{
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
	)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a missing domain")
	}
}

func TestCreate_MissingLogin(t *testing.T) {
	resp := createCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}),
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
	)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a missing login")
	}
}

func TestCreate_MissingPassword(t *testing.T) {
	resp := createCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}),
		map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, "example.com"),
			"login":  tftypes.NewValue(tftypes.String, "user"),
		},
		map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, "example.com"),
			"login":  tftypes.NewValue(tftypes.String, "user"),
		},
	)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when neither password source is set")
	}
}

func TestCreate_UnconfiguredClient(t *testing.T) {
	overrides := map[string]tftypes.Value{
		"domain":   tftypes.NewValue(tftypes.String, "example.com"),
		"login":    tftypes.NewValue(tftypes.String, "user"),
		"password": tftypes.NewValue(tftypes.String, "secret"),
	}

	resp := createCredential(t, nil, overrides, overrides)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the client is not configured")
	}
}

func TestCreate_CreateCredentialAPIError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	overrides := map[string]tftypes.Value{
		"domain":   tftypes.NewValue(tftypes.String, "example.com"),
		"login":    tftypes.NewValue(tftypes.String, "user"),
		"password": tftypes.NewValue(tftypes.String, "secret"),
	}

	resp := createCredential(t, client, overrides, overrides)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when CreateCredential fails")
	}
}

// The created_at lookup is best-effort: if the listing never catches up
// within the retry budget, Create must still succeed with created_at left
// null for Read to fill in later, not fail the apply.
func TestCreate_LookupNeverCatchesUp_LeavesCreatedAtNull(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
		}
	})

	overrides := map[string]tftypes.Value{
		"domain":   tftypes.NewValue(tftypes.String, "example.com"),
		"login":    tftypes.NewValue(tftypes.String, "user"),
		"password": tftypes.NewValue(tftypes.String, "secret"),
	}

	resp := createCredential(t, client, overrides, overrides)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func importCredential(t *testing.T, client *mailgun.Client, id string) *resource.ImportStateResponse {
	t.Helper()

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: SmtpCredentialResourceSchema()},
	}
	(&SmtpCredentialResource{client: client}).ImportState(context.Background(), resource.ImportStateRequest{ID: id}, resp)
	return resp
}

func TestImportState_InvalidID(t *testing.T) {
	resp := importCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}), "not-a-valid-id")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a malformed import ID")
	}
}

func TestImportState_UnconfiguredClient(t *testing.T) {
	resp := importCredential(t, nil, "example.com/user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the client is not configured")
	}
}

func TestImportState_ListFailureProducesError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := importCredential(t, client, "example.com/user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the credential listing fails")
	}
}

func TestImportState_CredentialNotFound(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
	})

	resp := importCredential(t, client, "example.com/gone-user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the credential does not exist")
	}
}

func TestImportState_Success(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(createdCredentialListResponse())
	})

	resp := importCredential(t, client, "example.com/user")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}
