package auth

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
)

type continuationStatusError struct {
	status     int
	message    string
	retryAfter *time.Duration
}

func (e continuationStatusError) Error() string {
	return e.message
}

func (e continuationStatusError) StatusCode() int {
	return e.status
}

func (e continuationStatusError) RetryAfter() *time.Duration {
	if e.retryAfter == nil {
		return nil
	}
	value := *e.retryAfter
	return &value
}

type continuationFailureExecutor struct {
	executeErr error
	stream     []cliproxyexecutor.StreamChunk
	calls      int
}

func (e *continuationFailureExecutor) Identifier() string { return "codex" }

func (e *continuationFailureExecutor) Execute(context.Context, *Auth, cliproxyexecutor.Request, cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	e.calls++
	return cliproxyexecutor.Response{}, e.executeErr
}

func (e *continuationFailureExecutor) ExecuteStream(context.Context, *Auth, cliproxyexecutor.Request, cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	e.calls++
	ch := make(chan cliproxyexecutor.StreamChunk, len(e.stream))
	for _, chunk := range e.stream {
		ch <- chunk
	}
	close(ch)
	return &cliproxyexecutor.StreamResult{Chunks: ch}, nil
}

func (e *continuationFailureExecutor) Refresh(_ context.Context, auth *Auth) (*Auth, error) {
	return auth, nil
}

func (e *continuationFailureExecutor) CountTokens(context.Context, *Auth, cliproxyexecutor.Request, cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	e.calls++
	return cliproxyexecutor.Response{}, e.executeErr
}

func (e *continuationFailureExecutor) HttpRequest(context.Context, *Auth, *http.Request) (*http.Response, error) {
	return nil, http.ErrNotSupported
}

func TestManagerExecute_PreviousResponseUsageLimitReturnsLocalCooldown(t *testing.T) {
	model := "gpt-5.5"
	retryAfter := 90 * time.Second
	upstreamErr := continuationStatusError{
		status:     http.StatusTooManyRequests,
		message:    `{"error":{"type":"invalid_request_error","message":"You've hit your usage limit."}}`,
		retryAfter: &retryAfter,
	}
	executor := &continuationFailureExecutor{executeErr: upstreamErr}
	manager := NewManager(nil, &RoundRobinSelector{}, nil)
	manager.RegisterExecutor(executor)

	registerContinuationTestAuths(t, manager, model, "continuation-exec-a", "continuation-exec-b")

	_, err := manager.Execute(context.Background(), []string{"codex"}, cliproxyexecutor.Request{Model: model}, continuationOptions())
	assertLocalContinuationCooldown(t, err, model)
	if executor.calls != 1 {
		t.Fatalf("executor calls = %d, want 1", executor.calls)
	}
}

func TestManagerExecute_PreviousResponseUsageLimitRedactionRequiresCodexContinuation(t *testing.T) {
	model := "gpt-5.5"
	upstreamErr := continuationStatusError{
		status:  http.StatusTooManyRequests,
		message: `{"error":{"message":"You've hit your usage limit."}}`,
	}
	manager := NewManager(nil, nil, nil)

	if got := manager.normalizeContinuationFailure([]string{"codex"}, model, cliproxyexecutor.Options{}, upstreamErr); got != upstreamErr {
		t.Fatalf("without previous_response_id error was normalized: %v", got)
	}
	if got := manager.normalizeContinuationFailure([]string{"gemini"}, model, continuationOptions(), upstreamErr); got != upstreamErr {
		t.Fatalf("non-codex error was normalized: %v", got)
	}
}

func TestManagerExecuteStream_PreviousResponseBootstrapUsageLimitReturnsLocalCooldown(t *testing.T) {
	model := "gpt-5.5"
	upstreamErr := continuationStatusError{
		status:  http.StatusTooManyRequests,
		message: `{"error":{"type":"invalid_request_error","message":"You've hit your usage limit."}}`,
	}
	executor := &continuationFailureExecutor{
		stream: []cliproxyexecutor.StreamChunk{{Err: upstreamErr}},
	}
	manager := NewManager(nil, &RoundRobinSelector{}, nil)
	manager.RegisterExecutor(executor)

	registerContinuationTestAuths(t, manager, model, "continuation-stream-bootstrap-a", "continuation-stream-bootstrap-b")

	result, err := manager.ExecuteStream(context.Background(), []string{"codex"}, cliproxyexecutor.Request{Model: model}, continuationOptions())
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if result == nil || result.Chunks == nil {
		t.Fatal("ExecuteStream() result/chunks is nil")
	}

	var gotErr error
	for chunk := range result.Chunks {
		if len(chunk.Payload) > 0 {
			t.Fatalf("unexpected payload = %q", string(chunk.Payload))
		}
		if chunk.Err != nil {
			gotErr = chunk.Err
		}
	}
	assertLocalContinuationCooldown(t, gotErr, model)
	if executor.calls != 1 {
		t.Fatalf("executor calls = %d, want 1", executor.calls)
	}
}

func TestManagerExecuteStream_PreviousResponseTerminalUsageLimitIsRedacted(t *testing.T) {
	model := "gpt-5.5"
	upstreamErr := continuationStatusError{
		status:  http.StatusTooManyRequests,
		message: `{"error":{"type":"invalid_request_error","message":"You've hit your usage limit."}}`,
	}
	executor := &continuationFailureExecutor{
		stream: []cliproxyexecutor.StreamChunk{
			{Payload: []byte("partial")},
			{Err: upstreamErr},
		},
	}
	manager := NewManager(nil, &RoundRobinSelector{}, nil)
	manager.RegisterExecutor(executor)

	registerContinuationTestAuths(t, manager, model, "continuation-stream-terminal")

	result, err := manager.ExecuteStream(context.Background(), []string{"codex"}, cliproxyexecutor.Request{Model: model}, continuationOptions())
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if result == nil || result.Chunks == nil {
		t.Fatal("ExecuteStream() result/chunks is nil")
	}

	var gotPayload []byte
	var gotErr error
	for chunk := range result.Chunks {
		if len(chunk.Payload) > 0 {
			gotPayload = append(gotPayload, chunk.Payload...)
		}
		if chunk.Err != nil {
			gotErr = chunk.Err
		}
	}
	if string(gotPayload) != "partial" {
		t.Fatalf("payload = %q, want partial", string(gotPayload))
	}
	assertLocalContinuationCooldown(t, gotErr, model)
}

func registerContinuationTestAuths(t *testing.T, manager *Manager, model string, authIDs ...string) {
	t.Helper()

	reg := registry.GetGlobalRegistry()
	for _, authID := range authIDs {
		reg.RegisterClient(authID, "codex", []*registry.ModelInfo{{ID: model}})
		if _, err := manager.Register(context.Background(), &Auth{ID: authID, Provider: "codex", Status: StatusActive}); err != nil {
			t.Fatalf("register auth %s: %v", authID, err)
		}
	}
	t.Cleanup(func() {
		for _, authID := range authIDs {
			reg.UnregisterClient(authID)
		}
	})
}

func continuationOptions() cliproxyexecutor.Options {
	return cliproxyexecutor.Options{
		OriginalRequest: []byte(`{"previous_response_id":"resp-1","input":[{"type":"message","role":"user","content":"continue"}]}`),
	}
}

func assertLocalContinuationCooldown(t *testing.T, err error, model string) {
	t.Helper()

	if err == nil {
		t.Fatal("error = nil, want local cooldown error")
	}
	cooldown, ok := err.(*modelCooldownError)
	if !ok {
		t.Fatalf("error type = %T, want *modelCooldownError: %v", err, err)
	}
	if cooldown.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", cooldown.StatusCode(), http.StatusTooManyRequests)
	}
	if cooldown.model != model {
		t.Fatalf("cooldown model = %q, want %q", cooldown.model, model)
	}
	body := cooldown.Error()
	if !strings.Contains(body, "model_cooldown") {
		t.Fatalf("local error missing model_cooldown: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "usage limit") {
		t.Fatalf("local error leaked upstream usage-limit text: %s", body)
	}
}
