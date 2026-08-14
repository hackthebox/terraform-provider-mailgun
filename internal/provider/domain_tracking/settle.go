// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	trackingSettleAttempts = 4
	trackingSettleDelay    = 250 * time.Millisecond
)

// settleTracking re-reads until the API reports what was just written. Mailgun
// seems to apply tracking writes asynchronously: the PUT returns 200 echoing the
// requested state while a read still reports the old one (#86). Giving up returns
// the last read, leaving it as drift for the next plan rather than failing apply.
func settleTracking(ctx context.Context, attempts int, delay time.Duration, plan *DomainTrackingModel,
	read func() (DomainTrackingModel, error),
) (DomainTrackingModel, error) {
	for attempt := 1; ; attempt++ {
		got, err := read()
		if err != nil {
			return got, err
		}
		if trackingSettled(plan, &got) || attempt >= attempts {
			return got, nil
		}

		if err := waitOrCancel(ctx, delay); err != nil {
			return got, err
		}
	}
}

func waitOrCancel(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// trackingSettled skips unknown plan values: nothing was written for them.
func trackingSettled(plan, got *DomainTrackingModel) bool {
	return boolSettled(plan.ClickActive, got.ClickActive) &&
		boolSettled(plan.OpenActive, got.OpenActive) &&
		boolSettled(plan.UnsubscribeActive, got.UnsubscribeActive)
}

func boolSettled(planned, got types.Bool) bool {
	return planned.IsUnknown() || planned.Equal(got)
}
