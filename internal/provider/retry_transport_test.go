// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

type fakeRoundTripper struct {
	responses []*http.Response
	err       error
	calls     int
	bodies    []string
}

func (f *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		f.bodies = append(f.bodies, string(body))
	}
	if f.err != nil {
		return nil, f.err
	}
	resp := f.responses[min(f.calls-1, len(f.responses)-1)]
	return resp, nil
}

func response(status int, retryAfter string) *http.Response {
	header := http.Header{}
	if retryAfter != "" {
		header.Set("Retry-After", retryAfter)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(""))}
}

func newTestTransport(next http.RoundTripper, recorded *[]time.Duration) rateLimitRetryTransport {
	return rateLimitRetryTransport{
		next:     next,
		attempts: 3,
		wait: func(_ context.Context, d time.Duration) error {
			*recorded = append(*recorded, d)
			return nil
		},
	}
}

func TestRoundTripPassesThroughNonRateLimited(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{response(http.StatusOK, "")}}
	var waits []time.Duration

	resp, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1", fake.calls)
	}
	if len(waits) != 0 {
		t.Errorf("waited %v, want no waiting", waits)
	}
}

func TestRoundTripRetriesRateLimitedRequest(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{
		response(http.StatusTooManyRequests, ""),
		response(http.StatusOK, ""),
	}}
	var waits []time.Duration

	resp, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 after retry", resp.StatusCode)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2", fake.calls)
	}
}

func TestRoundTripHonoursRetryAfter(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{
		response(http.StatusTooManyRequests, "7"),
		response(http.StatusOK, ""),
	}}
	var waits []time.Duration

	if _, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(waits) != 1 || waits[0] != 7*time.Second {
		t.Errorf("waits = %v, want [7s]", waits)
	}
}

func TestRoundTripGivesUpAfterMaxAttempts(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{response(http.StatusTooManyRequests, "")}}
	var waits []time.Duration

	resp, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want the final 429 surfaced to the caller", resp.StatusCode)
	}
	if fake.calls != 4 {
		t.Errorf("calls = %d, want 4 (initial + 3 retries)", fake.calls)
	}
}

func TestRoundTripReplaysRequestBodyOnRetry(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{
		response(http.StatusTooManyRequests, ""),
		response(http.StatusOK, ""),
	}}
	var waits []time.Duration

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://example.invalid/v3/routes", strings.NewReader("name=value"))
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	if _, err := newTestTransport(fake, &waits).RoundTrip(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.bodies) != 2 {
		t.Fatalf("recorded %d bodies, want 2", len(fake.bodies))
	}
	if fake.bodies[0] != "name=value" || fake.bodies[1] != "name=value" {
		t.Errorf("bodies = %v, want the payload replayed verbatim", fake.bodies)
	}
}

func TestRoundTripDoesNotRetryUnreplayableBody(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{response(http.StatusTooManyRequests, "")}}
	var waits []time.Duration

	req := httpGet(t)
	req.Method = http.MethodPost
	req.Body = io.NopCloser(strings.NewReader("streamed"))
	req.GetBody = nil

	resp, err := newTestTransport(fake, &waits).RoundTrip(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want the 429 surfaced without retry", resp.StatusCode)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1 (body cannot be replayed)", fake.calls)
	}
}

func TestRoundTripSurfacesTransportError(t *testing.T) {
	wantErr := errors.New("dial failed")
	fake := &fakeRoundTripper{err: wantErr}
	var waits []time.Duration

	if _, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t)); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func TestRoundTripAbortsWhenWaitFails(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{response(http.StatusTooManyRequests, "")}}
	transport := rateLimitRetryTransport{
		next:     fake,
		attempts: 3,
		wait:     func(context.Context, time.Duration) error { return context.Canceled },
	}

	if _, err := transport.RoundTrip(httpGet(t)); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestRetryAfterDelay(t *testing.T) {
	fallback := 2 * time.Second
	tests := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{"absent", "", fallback},
		{"seconds", "5", 5 * time.Second},
		{"zero", "0", fallback},
		{"negative", "-3", fallback},
		{"unparseable", "later", fallback},
		{"one second", "1", time.Second},
		{"capped", "9999", 30 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := retryAfterDelay(tt.header, fallback); got != tt.want {
				t.Errorf("retryAfterDelay(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

func httpGet(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/v3/routes", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	return req
}

func TestSleepWithContextCompletes(t *testing.T) {
	if err := sleepWithContext(context.Background(), time.Millisecond); err != nil {
		t.Errorf("err = %v, want nil", err)
	}
}

func TestSleepWithContextAbortsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := sleepWithContext(ctx, time.Minute); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestNewRateLimitRetryTransportDefaults(t *testing.T) {
	transport := newRateLimitRetryTransport()

	if transport.attempts != 3 {
		t.Errorf("attempts = %d, want 3", transport.attempts)
	}
	if transport.next != http.DefaultTransport {
		t.Error("next should default to http.DefaultTransport")
	}
	if transport.wait == nil {
		t.Error("wait must be set so RoundTrip can back off")
	}
}

type countingBody struct {
	io.Reader
	closes *int
}

func (b countingBody) Close() error {
	*b.closes++
	return nil
}

func TestRoundTripBacksOffExponentiallyWithoutRetryAfter(t *testing.T) {
	fake := &fakeRoundTripper{responses: []*http.Response{
		response(http.StatusTooManyRequests, ""),
		response(http.StatusTooManyRequests, ""),
		response(http.StatusOK, ""),
	}}
	var waits []time.Duration

	if _, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(waits) != 2 || waits[0] != 2*time.Second || waits[1] != 4*time.Second {
		t.Errorf("waits = %v, want [2s 4s] growing per attempt", waits)
	}
}

func TestRoundTripClosesDiscardedResponseBodies(t *testing.T) {
	var closes int
	rateLimited := response(http.StatusTooManyRequests, "")
	rateLimited.Body = countingBody{Reader: strings.NewReader(""), closes: &closes}

	fake := &fakeRoundTripper{responses: []*http.Response{rateLimited, response(http.StatusOK, "")}}
	var waits []time.Duration

	if _, err := newTestTransport(fake, &waits).RoundTrip(httpGet(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if closes != 1 {
		t.Errorf("closes = %d, want 1 so the discarded 429 body does not leak", closes)
	}
}

func TestRoundTripPassesRequestContextToWait(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(context.Background(), marker{}, "carried")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.invalid/v3/routes", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	fake := &fakeRoundTripper{responses: []*http.Response{
		response(http.StatusTooManyRequests, ""),
		response(http.StatusOK, ""),
	}}
	var seen context.Context
	transport := rateLimitRetryTransport{
		next:     fake,
		attempts: 3,
		wait: func(c context.Context, _ time.Duration) error {
			seen = c
			return nil
		},
	}

	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seen == nil || seen.Value(marker{}) != "carried" {
		t.Error("wait must receive the request context so cancellation propagates")
	}
}

func TestConfigureInstallsRetryTransport(t *testing.T) {
	ctx := context.Background()

	var schemaResp provider.SchemaResponse
	(&mailgunProvider{}).Schema(ctx, provider.SchemaRequest{}, &schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("provider schema type is not an object")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["api_key"] = tftypes.NewValue(tftypes.String, "key-test")

	var resp provider.ConfigureResponse
	(&mailgunProvider{}).Configure(ctx, provider.ConfigureRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, attrs), Schema: schemaResp.Schema},
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	client, ok := resp.ResourceData.(*mailgun.Client)
	if !ok {
		t.Fatalf("ResourceData = %T, want *mailgun.Client", resp.ResourceData)
	}
	if _, ok := client.HTTPClient().Transport.(rateLimitRetryTransport); !ok {
		t.Errorf("transport = %T, want rateLimitRetryTransport so 429s are retried", client.HTTPClient().Transport)
	}
}

func TestRoundTripReusesConnectionAcrossRetries(t *testing.T) {
	var newConns int32
	var calls int32

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"message":"Too many requests"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt32(&newConns, 1)
		}
	}
	server.Start()
	defer server.Close()

	transport := rateLimitRetryTransport{
		next:     &http.Transport{},
		attempts: 3,
		wait:     func(context.Context, time.Duration) error { return nil },
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&newConns); got != 1 {
		t.Errorf("new connections = %d, want 1: the discarded 429 body must be drained or net/http drops the connection", got)
	}
}
