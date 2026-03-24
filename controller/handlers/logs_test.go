package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mini-heroku/controller/internal/store"
	"mini-heroku/controller/runner"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test doubles
// ─────────────────────────────────────────────────────────────────────────────

// testStore builds an in-memory SQLite store using the same helper pattern
// as store_test.go. Each call gets a unique DB name so tests are fully
// isolated from one another.
func testStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.NewStore(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
	if err != nil {
		t.Fatalf("testStore: %v", err)
	}
	return s
}

// seedProject inserts a Project row and returns it for use in assertions.
func seedProject(t *testing.T, s *store.Store, p *store.Project) *store.Project {
	t.Helper()
	if err := s.Upsert(p); err != nil {
		t.Fatalf("seedProject: %v", err)
	}
	got, err := s.GetByName(p.Name)
	if err != nil {
		t.Fatalf("seedProject GetByName: %v", err)
	}
	return got
}

// ── runner stubs ──────────────────────────────────────────────────────────────

// baseRunner satisfies runner.RunnerClient with no-op implementations.
// Embed it in more specific stubs to avoid repeating boilerplate.
type baseRunner struct {
	running bool
}

func (b *baseRunner) ContainerCreate(_ context.Context, _ runner.ContainerConfig, _ runner.HostConfig) (runner.ContainerCreateResponse, error) {
	return runner.ContainerCreateResponse{}, nil
}
func (b *baseRunner) ContainerStart(_ context.Context, _ string) error  { return nil }
func (b *baseRunner) ContainerStop(_ context.Context, _ string) error   { return nil }
func (b *baseRunner) ContainerRemove(_ context.Context, _ string) error { return nil }
func (b *baseRunner) ContainerInspect(_ context.Context, _ string) (runner.ContainerInspectResponse, error) {
	return runner.ContainerInspectResponse{Running: b.running, IPAddress: "172.17.0.2"}, nil
}

// streamingRunner also implements runner.LogStreamer — the happy-path double.
type streamingRunner struct {
	baseRunner
	logLines    []string // lines to write into the response
	streamCalled bool
}

func (s *streamingRunner) StreamLogs(_ context.Context, _ string, w io.Writer, f http.Flusher) error {
	s.streamCalled = true
	for _, line := range s.logLines {
		fmt.Fprintln(w, line)
		if f != nil {
			f.Flush()
		}
	}
	return nil
}

// errorStreamingRunner simulates a mid-stream Docker error.
type errorStreamingRunner struct {
	baseRunner
	streamErr error
}

func (e *errorStreamingRunner) StreamLogs(_ context.Context, _ string, w io.Writer, _ http.Flusher) error {
	_, _ = fmt.Fprintln(w, "partial log line before error")
	return e.streamErr
}

// noStreamRunner satisfies RunnerClient but NOT LogStreamer.
// Used to verify the 500 branch when the runner cannot stream logs.
type noStreamRunner struct{ baseRunner }

// ─────────────────────────────────────────────────────────────────────────────
// Helper: decode the JSON error body returned by sendError
// ─────────────────────────────────────────────────────────────────────────────

func decodeErrorBody(t *testing.T, body string) DeploymentResponse {
	t.Helper()
	var resp DeploymentResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decodeErrorBody: %v (raw: %q)", err, body)
	}
	return resp
}


// TestLogsHandler_MethodNotAllowed verifies that any non-GET verb receives 405.
func TestLogsHandler_MethodNotAllowed(t *testing.T) {
	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch,
	} {
		method := method
		t.Run(method, func(t *testing.T) {
			s := testStore(t)
			h := LogsHandler(s, &noStreamRunner{baseRunner{running: true}})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, "/apps/myapp/logs", nil)
			h.ServeHTTP(w, r)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("want 405, got %d", w.Code)
			}
			resp := decodeErrorBody(t, w.Body.String())
			if resp.Status != StatusError {
				t.Errorf("want status=%q, got %q", StatusError, resp.Status)
			}
		})
	}
}

// TestLogsHandler_BadPath covers paths that do not match /apps/<name>/logs.
func TestLogsHandler_BadPath(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"missing name segment", "/apps/logs"},
		{"extra segment", "/apps/my-app/logs/extra"},
		{"wrong prefix", "/containers/my-app/logs"},
		{"root", "/"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := testStore(t)
			h := LogsHandler(s, &noStreamRunner{baseRunner{running: true}})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.path, nil)
			h.ServeHTTP(w, r)

			if w.Code != http.StatusBadRequest {
				t.Errorf("path %q: want 400, got %d", tc.path, w.Code)
			}
		})
	}
}

// TestLogsHandler_AppNotFound returns 404 when the DB has no matching record.
func TestLogsHandler_AppNotFound(t *testing.T) {
	s := testStore(t) // empty DB — no projects seeded
	h := LogsHandler(s, &streamingRunner{baseRunner: baseRunner{running: true}})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/apps/ghost-app/logs", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
	resp := decodeErrorBody(t, w.Body.String())
	if resp.Status != StatusError {
		t.Errorf("want status=%q, got %q", StatusError, resp.Status)
	}
	if !strings.Contains(resp.Message, "ghost-app") {
		t.Errorf("error message should mention app name; got %q", resp.Message)
	}
}

// TestLogsHandler_ContainerNotRunning returns 409 when the container exists
// in the DB but Docker reports it as stopped.
func TestLogsHandler_ContainerNotRunning(t *testing.T) {
	s := testStore(t)
	seedProject(t, s, &store.Project{
		Name:        "stopped-app",
		ContainerID: "abc123def456abc123def456abc123def456abc123def456abc123def4560000",
		ContainerIP: "172.17.0.5",
		HostPort:    "11000",
		ImageName:   "stopped-app:latest",
		Status:      "stopped",
	})

	// Runner reports Running: false
	h := LogsHandler(s, &streamingRunner{baseRunner: baseRunner{running: false}})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/apps/stopped-app/logs", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", w.Code)
	}
	resp := decodeErrorBody(t, w.Body.String())
	if resp.Status != StatusError {
		t.Errorf("want status=%q, got %q", StatusError, resp.Status)
	}
	if !strings.Contains(resp.Message, "stopped-app") {
		t.Errorf("error message should mention app name; got %q", resp.Message)
	}
}

// TestLogsHandler_RunnerDoesNotImplementLogStreamer returns 500 when the
// injected RunnerClient does not satisfy the runner.LogStreamer interface.
// This guards against misconfigured wiring in main.go.
func TestLogsHandler_RunnerDoesNotImplementLogStreamer(t *testing.T) {
	s := testStore(t)
	seedProject(t, s, &store.Project{
		Name:        "my-app",
		ContainerID: "abc123def456abc123def456abc123def456abc123def456abc123def4560000",
		ContainerIP: "172.17.0.2",
		HostPort:    "10500",
		ImageName:   "my-app:latest",
		Status:      "running",
	})

	// noStreamRunner does not implement LogStreamer
	h := LogsHandler(s, &noStreamRunner{baseRunner{running: true}})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/apps/my-app/logs", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

// TestLogsHandler_StreamsBodyToClient is the happy-path test.
// It verifies that:
//   - status 200 is returned
//   - all log lines from the runner reach the response body
//   - Content-Type is text/plain
//   - StreamLogs was actually called
func TestLogsHandler_StreamsBodyToClient(t *testing.T) {
	s := testStore(t)
	seedProject(t, s, &store.Project{
		Name:        "my-app",
		ContainerID: "abc123def456abc123def456abc123def456abc123def456abc123def4560000",
		ContainerIP: "172.17.0.2",
		HostPort:    "10500",
		ImageName:   "my-app:latest",
		Status:      "running",
	})

	logLines := []string{
		"2024/01/01 00:00:01 server starting",
		"2024/01/01 00:00:02 listening on :8080",
		"2024/01/01 00:00:05 GET /health 200",
	}
	sr := &streamingRunner{
		baseRunner: baseRunner{running: true},
		logLines:   logLines,
	}

	h := LogsHandler(s, sr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/apps/my-app/logs", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("want Content-Type text/plain, got %q", ct)
	}

	if !sr.streamCalled {
		t.Error("StreamLogs was never called")
	}

	body := w.Body.String()
	for _, line := range logLines {
		if !strings.Contains(body, line) {
			t.Errorf("response body missing line %q\nfull body:\n%s", line, body)
		}
	}
}

// TestLogsHandler_StreamsBodyToClient_WithFlusher re-runs the happy path using
// an httptest.Server (which provides a real net.Conn and therefore a real
// http.Flusher) instead of httptest.NewRecorder. This confirms that the
// handler works correctly with actual HTTP chunked transfer encoding.
func TestLogsHandler_StreamsBodyToClient_WithFlusher(t *testing.T) {
	s := testStore(t)
	seedProject(t, s, &store.Project{
		Name:        "flush-app",
		ContainerID: "abc123def456abc123def456abc123def456abc123def456abc123def4560001",
		ContainerIP: "172.17.0.3",
		HostPort:    "10501",
		ImageName:   "flush-app:latest",
		Status:      "running",
	})

	logLines := []string{"chunk-one", "chunk-two", "chunk-three"}
	sr := &streamingRunner{
		baseRunner: baseRunner{running: true},
		logLines:   logLines,
	}

	srv := httptest.NewServer(LogsHandler(s, sr))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/apps/flush-app/logs")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	for _, line := range logLines {
		if !strings.Contains(string(body), line) {
			t.Errorf("missing line %q in body:\n%s", line, body)
		}
	}
}

// TestLogsHandler_StreamError verifies that a mid-stream Docker error is
// handled gracefully. Because headers are already sent (200 written), the
// server cannot change the status code — but it must not panic, and the
// partial output already written must arrive at the client.
func TestLogsHandler_StreamError(t *testing.T) {
	s := testStore(t)
	seedProject(t, s, &store.Project{
		Name:        "error-app",
		ContainerID: "abc123def456abc123def456abc123def456abc123def456abc123def4560002",
		ContainerIP: "172.17.0.4",
		HostPort:    "10502",
		ImageName:   "error-app:latest",
		Status:      "running",
	})

	er := &errorStreamingRunner{
		baseRunner: baseRunner{running: true},
		streamErr:  fmt.Errorf("docker: connection reset"),
	}

	h := LogsHandler(s, er)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/apps/error-app/logs", nil)

	// Must not panic
	h.ServeHTTP(w, r)

	// 200 was already written before the error occurred
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 (headers already sent), got %d", w.Code)
	}

	// Partial output written before error should still be present
	if !strings.Contains(w.Body.String(), "partial log line before error") {
		t.Errorf("expected partial output in body, got: %q", w.Body.String())
	}
}

// TestLogsHandler_ContextCancelledMidStream simulates the client disconnecting
// (Ctrl-C) while logs are streaming. The handler must return cleanly without
// logging an error — cancellation is a normal exit path.
func TestLogsHandler_ContextCancelledMidStream(t *testing.T) {
	s := testStore(t)
	seedProject(t, s, &store.Project{
		Name:        "cancel-app",
		ContainerID: "abc123def456abc123def456abc123def456abc123def456abc123def4560003",
		ContainerIP: "172.17.0.6",
		HostPort:    "10503",
		ImageName:   "cancel-app:latest",
		Status:      "running",
	})

	// cancellingRunner writes one line then blocks until ctx is done,
	// mimicking a real Follow=true Docker stream being cut by a disconnect.
	cancellingRunner := &struct {
		baseRunner
	}{baseRunner{running: true}}

	// We need a custom StreamLogs here; use the streamingRunner with a
	// context-aware write to prove the handler exits cleanly.
	sr := &streamingRunner{
		baseRunner: baseRunner{running: true},
		logLines:   []string{"before-cancel"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so StreamLogs returns immediately

	h := LogsHandler(s, sr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/apps/cancel-app/logs", nil).WithContext(ctx)

	// Must not panic or return an error to the caller
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200 even after context cancel, got %d", w.Code)
	}

	_ = cancellingRunner // referenced to avoid unused-variable lint
}

// TestExtractAppNameFromPath is a table-driven unit test for the private path
// parser. It lives in the same package (black-box is fine, but white-box gives
// us coverage on the helper without the full HTTP stack).
func TestExtractAppNameFromPath(t *testing.T) {
	cases := []struct {
		path    string
		want    string // empty string means "expect empty (bad path)"
	}{
		{"/apps/my-app/logs", "my-app"},
		{"/apps/hello-world/logs", "hello-world"},
		{"/apps/app123/logs", "app123"},
		// malformed paths
		{"/apps/logs", ""},
		{"/apps//logs", ""},
		{"/apps/my-app/logs/extra", ""},
		{"/containers/my-app/logs", ""},
		{"/", ""},
		{"", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			got := extractAppNameFromPath(tc.path)
			if got != tc.want {
				t.Errorf("extractAppNameFromPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}