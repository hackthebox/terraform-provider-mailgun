// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTrackingSettledIgnoresUnknownPlanValues(t *testing.T) {
	plan := DomainTrackingModel{
		ClickActive:       types.BoolUnknown(),
		OpenActive:        types.BoolUnknown(),
		UnsubscribeActive: types.BoolUnknown(),
	}
	got := DomainTrackingModel{
		ClickActive:       types.BoolValue(true),
		OpenActive:        types.BoolValue(false),
		UnsubscribeActive: types.BoolValue(true),
	}

	if !trackingSettled(&plan, &got) {
		t.Error("nothing was written for unknown plan values, so any read counts as settled")
	}
}

func TestTrackingSettledDetectsAStaleRead(t *testing.T) {
	plan := DomainTrackingModel{
		ClickActive:       types.BoolValue(true),
		OpenActive:        types.BoolValue(true),
		UnsubscribeActive: types.BoolValue(true),
	}
	stale := DomainTrackingModel{
		ClickActive:       types.BoolValue(true),
		OpenActive:        types.BoolValue(true),
		UnsubscribeActive: types.BoolValue(false),
	}

	if trackingSettled(&plan, &stale) {
		t.Error("unsubscribe_active still reads false, which is not settled")
	}
}

func TestTrackingSettledAcceptsAMatchingRead(t *testing.T) {
	plan := DomainTrackingModel{
		ClickActive:       types.BoolValue(false),
		OpenActive:        types.BoolValue(true),
		UnsubscribeActive: types.BoolValue(true),
	}
	got := plan

	if !trackingSettled(&plan, &got) {
		t.Error("a read matching every pinned value is settled")
	}
}

func TestSettleTrackingReturnsFirstMatchingRead(t *testing.T) {
	plan := DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}
	reads := 0

	got, err := settleTracking(context.Background(), 4, 0, &plan, func() (DomainTrackingModel, error) {
		reads++
		return DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.UnsubscribeActive.ValueBool() {
		t.Error("expected the settled read to be returned")
	}
	if reads != 1 {
		t.Errorf("reads = %d, want 1 when the first read already agrees", reads)
	}
}

func TestSettleTrackingRetriesUntilTheWriteIsVisible(t *testing.T) {
	plan := DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}
	reads := 0

	got, err := settleTracking(context.Background(), 4, 0, &plan, func() (DomainTrackingModel, error) {
		reads++
		if reads < 3 {
			return DomainTrackingModel{UnsubscribeActive: types.BoolValue(false)}, nil
		}
		return DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.UnsubscribeActive.ValueBool() {
		t.Error("expected the eventually-visible value")
	}
	if reads != 3 {
		t.Errorf("reads = %d, want 3", reads)
	}
}

func TestSettleTrackingGivesUpAndReturnsTheLastRead(t *testing.T) {
	plan := DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}
	reads := 0

	got, err := settleTracking(context.Background(), 3, 0, &plan, func() (DomainTrackingModel, error) {
		reads++
		return DomainTrackingModel{UnsubscribeActive: types.BoolValue(false)}, nil
	})

	if err != nil {
		t.Fatalf("a write that never becomes visible is drift, not an apply failure: %v", err)
	}
	if got.UnsubscribeActive.ValueBool() {
		t.Error("expected the last observed value, not the planned one")
	}
	if reads != 3 {
		t.Errorf("reads = %d, want 3", reads)
	}
}

func TestSettleTrackingPropagatesReadErrors(t *testing.T) {
	plan := DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}
	wantErr := errors.New("read failed")

	if _, err := settleTracking(context.Background(), 4, 0, &plan, func() (DomainTrackingModel, error) {
		return DomainTrackingModel{}, wantErr
	}); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func TestSettleTrackingAbortsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	plan := DomainTrackingModel{UnsubscribeActive: types.BoolValue(true)}
	reads := 0

	_, err := settleTracking(ctx, 4, time.Minute, &plan, func() (DomainTrackingModel, error) {
		reads++
		return DomainTrackingModel{UnsubscribeActive: types.BoolValue(false)}, nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if reads != 1 {
		t.Errorf("reads = %d, want 1 before the context cancels the wait", reads)
	}
}
