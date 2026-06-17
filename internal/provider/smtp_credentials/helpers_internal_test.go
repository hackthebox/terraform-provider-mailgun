// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func TestFindEventuallyReturnsFirstSuccess(t *testing.T) {
	calls := 0
	want := &mtypes.Credential{Login: "user@example.com"}

	got, err := findEventually(context.Background(), 4, 0, func() (*mtypes.Credential, error) {
		calls++
		return want, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("credential = %v, want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (no retry once the listing succeeds)", calls)
	}
}

func TestFindEventuallyRetriesUntilTheListingCatchesUp(t *testing.T) {
	calls := 0
	want := &mtypes.Credential{Login: "user@example.com"}

	got, err := findEventually(context.Background(), 4, 0, func() (*mtypes.Credential, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("credential not found")
		}
		return want, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("credential = %v, want %v", got, want)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestFindEventuallyGivesUpAfterAttempts(t *testing.T) {
	calls := 0
	wantErr := errors.New("credential not found")

	_, err := findEventually(context.Background(), 2, 0, func() (*mtypes.Credential, error) {
		calls++
		return nil, wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (the attempt budget, not the package default)", calls)
	}
}

func TestFindEventuallyAbortsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	_, err := findEventually(ctx, 4, time.Minute, func() (*mtypes.Credential, error) {
		calls++
		return nil, errors.New("credential not found")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 before the context cancels the wait", calls)
	}
}

func TestFindEventuallyDefaultsToFourAttempts(t *testing.T) {
	calls := 0

	_, err := findEventually(context.Background(), createdAtLookupAttempts, 0, func() (*mtypes.Credential, error) {
		calls++
		return nil, errors.New("credential not found")
	})

	if err == nil {
		t.Fatal("expected the lookup to give up")
	}
	if calls != 4 {
		t.Errorf("calls = %d, want 4: the package default budget", calls)
	}
}

func TestCreatedAtLookupActuallyBacksOff(t *testing.T) {
	if createdAtLookupDelay <= 0 {
		t.Fatalf("createdAtLookupDelay = %v; a zero delay retries the listing with no time to catch up", createdAtLookupDelay)
	}
	if createdAtLookupDelay > time.Second {
		t.Errorf("createdAtLookupDelay = %v; too long to sit in a create", createdAtLookupDelay)
	}
}

func TestCreatedAtLookupBudgetCoversItsAttempts(t *testing.T) {
	if createdAtLookupBudget <= 0 {
		t.Fatalf("createdAtLookupBudget = %v; a non-positive budget expires before the first lookup runs", createdAtLookupBudget)
	}

	// The budget has to outlast the waits it wraps, or later attempts never happen.
	if minimum := time.Duration(createdAtLookupAttempts-1) * createdAtLookupDelay; createdAtLookupBudget <= minimum {
		t.Errorf("createdAtLookupBudget = %v, must exceed the %v spent waiting between attempts", createdAtLookupBudget, minimum)
	}
}

func TestPasswordForCreate(t *testing.T) {
	tests := []struct {
		name       string
		passwordWO types.String
		legacy     types.String
		wantPass   string
		wantOK     bool
	}{
		{"write-only preferred", types.StringValue("wo-secret"), types.StringValue("legacy"), "wo-secret", true},
		{"legacy when no wo", types.StringNull(), types.StringValue("legacy"), "legacy", true},
		{"neither set", types.StringNull(), types.StringNull(), "", false},
		{"legacy unknown ignored", types.StringNull(), types.StringUnknown(), "", false},
		{"wo unknown falls back to legacy", types.StringUnknown(), types.StringValue("legacy"), "legacy", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPass, gotOK := passwordForCreate(tt.passwordWO, tt.legacy)
			if gotPass != tt.wantPass || gotOK != tt.wantOK {
				t.Errorf("passwordForCreate() = (%q, %v), want (%q, %v)", gotPass, gotOK, tt.wantPass, tt.wantOK)
			}
		})
	}
}

func TestWriteOnlyRotationRequested(t *testing.T) {
	tests := []struct {
		name  string
		plan  types.Int64
		state types.Int64
		want  bool
	}{
		{"version bumped", types.Int64Value(2), types.Int64Value(1), true},
		{"version unchanged", types.Int64Value(1), types.Int64Value(1), false},
		{"first set from null state", types.Int64Value(1), types.Int64Null(), true},
		{"no version in plan", types.Int64Null(), types.Int64Null(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeOnlyRotationRequested(tt.plan, tt.state); got != tt.want {
				t.Errorf("writeOnlyRotationRequested() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveUpdatePassword(t *testing.T) {
	tests := []struct {
		name         string
		passwordWO   types.String
		planPW       types.String
		planVersion  types.Int64
		stateVersion types.Int64
		wantPW       string
		wantRotate   bool
		wantErr      string
	}{
		{
			name:         "write-only with version bump rotates",
			passwordWO:   types.StringValue("wo-secret"),
			planPW:       types.StringNull(),
			planVersion:  types.Int64Value(2),
			stateVersion: types.Int64Value(1),
			wantPW:       "wo-secret",
			wantRotate:   true,
			wantErr:      "",
		},
		{
			name:         "write-only without version bump skips rotation",
			passwordWO:   types.StringValue("wo-secret"),
			planPW:       types.StringNull(),
			planVersion:  types.Int64Value(1),
			stateVersion: types.Int64Value(1),
			wantPW:       "",
			wantRotate:   false,
			wantErr:      "",
		},
		{
			name:         "legacy null preserves imported state",
			passwordWO:   types.StringNull(),
			planPW:       types.StringNull(),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "",
			wantRotate:   false,
			wantErr:      "",
		},
		{
			name:         "legacy empty string is an error",
			passwordWO:   types.StringNull(),
			planPW:       types.StringValue(""),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "",
			wantRotate:   false,
			wantErr:      "Invalid Password",
		},
		{
			name:         "legacy non-empty rotates",
			passwordWO:   types.StringNull(),
			planPW:       types.StringValue("newpass"),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "newpass",
			wantRotate:   true,
			wantErr:      "",
		},
		{
			name:         "write-only unknown falls through to legacy",
			passwordWO:   types.StringUnknown(),
			planPW:       types.StringValue("legacy"),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "legacy",
			wantRotate:   true,
			wantErr:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPW, gotRotate, gotErr := resolveUpdatePassword(tt.passwordWO, tt.planPW, tt.planVersion, tt.stateVersion)
			if gotPW != tt.wantPW || gotRotate != tt.wantRotate || gotErr != tt.wantErr {
				t.Errorf("resolveUpdatePassword() = (%q, %v, %q), want (%q, %v, %q)",
					gotPW, gotRotate, gotErr, tt.wantPW, tt.wantRotate, tt.wantErr)
			}
		})
	}
}
