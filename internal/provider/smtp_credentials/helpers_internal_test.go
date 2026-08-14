// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"errors"
	"testing"
	"time"

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
