// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

const (
	rateLimitRetryAttempts = 3
	baseRetryBackoff       = 2 * time.Second
	maxRetryAfter          = 30 * time.Second
)

// rateLimitRetryTransport retries 429s. Mailgun's routes endpoints have a low
// ceiling that a moderate configuration reaches on its own, and unretried it
// reaches practitioners as a hard apply failure.
type rateLimitRetryTransport struct {
	next     http.RoundTripper
	attempts int
	wait     func(context.Context, time.Duration) error
}

func newRateLimitRetryTransport() rateLimitRetryTransport {
	return rateLimitRetryTransport{
		next:     http.DefaultTransport,
		attempts: rateLimitRetryAttempts,
		wait:     sleepWithContext,
	}
}

func (t rateLimitRetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := t.next.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if !t.retryable(resp, attempt) {
			return resp, nil
		}

		replay, replayable := rewindBody(req)
		if !replayable {
			return resp, nil
		}

		if err := t.pause(req.Context(), resp, attempt); err != nil {
			return nil, err
		}
		req = replay
	}
}

func (t rateLimitRetryTransport) retryable(resp *http.Response, attempt int) bool {
	return resp.StatusCode == http.StatusTooManyRequests && attempt < t.attempts
}

func (t rateLimitRetryTransport) pause(ctx context.Context, resp *http.Response, attempt int) error {
	delay := retryAfterDelay(resp.Header.Get("Retry-After"), baseRetryBackoff<<attempt)
	resp.Body.Close()
	return t.wait(ctx, delay)
}

// retryAfterDelay reads Retry-After, which Mailgun sends as whole seconds.
func retryAfterDelay(header string, fallback time.Duration) time.Duration {
	seconds, err := strconv.Atoi(header)
	if err != nil || seconds <= 0 {
		return fallback
	}
	if delay := time.Duration(seconds) * time.Second; delay < maxRetryAfter {
		return delay
	}
	return maxRetryAfter
}

func rewindBody(req *http.Request) (*http.Request, bool) {
	if req.Body == nil {
		return req, true
	}
	if req.GetBody == nil {
		return nil, false
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, false
	}

	replay := req.Clone(req.Context())
	replay.Body = body
	return replay, true
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
