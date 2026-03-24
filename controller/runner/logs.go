package runner

import (
	"context"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

type LogStreamer interface {
	StreamLogs(ctx context.Context, containerID string, w io.Writer, flusher http.Flusher) error
}

func (r *RealRunnerClient) StreamLogs(ctx context.Context, containerID string, w io.Writer, flusher http.Flusher) error {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	logReader, err := r.client.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return err
	}
	defer logReader.Close()

	dst := &flushedWriter{w: w, flusher: flusher}

	_, err = stdcopy.StdCopy(dst, dst, logReader)

	if ctx.Err() != nil {
		return nil
	}
	return err
}

type flushedWriter struct {
	w       io.Writer
	flusher http.Flusher
}

func (fw *flushedWriter) Write(p []byte) (int, error) {
	n, err := fw.w.Write(p)
	if err != nil {
		return n, err
	}
	if fw.flusher != nil {
		fw.flusher.Flush()
	}
	return n, nil
}
