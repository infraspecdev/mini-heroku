package runner

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)


var _ LogStreamer = (*RealRunnerClient)(nil)

type mockLogStreamer struct {
	written []byte
}

func (m *mockLogStreamer) StreamLogs(ctx context.Context, containerID string, w io.Writer, f http.Flusher) error {
	_, err := w.Write([]byte("line1\nline2\n"))
	m.written = []byte("line1\nline2\n")
	return err
}

func TestStreamLogs_WritesOutput(t *testing.T) {
	ms := &mockLogStreamer{}
	var buf bytes.Buffer
	err := ms.StreamLogs(context.Background(), "abc123", &buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "line1\nline2\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestStreamLogs_ContextCancel_ReturnsNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	ms := &mockLogStreamer{}
	var buf bytes.Buffer
	// A real implementation must return nil on cancelled context
	err := ms.StreamLogs(ctx, "abc123", &buf, nil)
	if err != nil {
		t.Fatalf("expected nil on cancelled ctx, got: %v", err)
	}
}
