package builder

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

type MockDockerClient struct {
	BuildCalled  bool
	BuildContext []byte
	BuildOptions ImageBuildOptions
	ReturnError  error
	ReturnBody   string
}

func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options ImageBuildOptions) (ImageBuildResponse, error) {
	m.BuildCalled = true
	m.BuildOptions = options

	var err error
	m.BuildContext, err = io.ReadAll(buildContext)
	if err != nil {
		return ImageBuildResponse{}, err
	}

	if m.ReturnError != nil {
		return ImageBuildResponse{}, m.ReturnError
	}

	body := m.ReturnBody
	if body == "" {
		body = `{"stream":"Successfully built abc123"}`
	}

	return ImageBuildResponse{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}, nil
}

func TestBuildImage(t *testing.T) {
	mockClient := &MockDockerClient{}
	tarball := []byte("fake-tar")

	imageID, err := BuildImage(mockClient, bytes.NewReader(tarball), "test-app")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockClient.BuildCalled {
		t.Fatal("ImageBuild was not called")
	}

	expectedTag := "test-app:latest"

	if imageID != expectedTag {
		t.Fatalf("expected is %s, got %s", expectedTag, imageID)
	}
}

func TestBuildImage_DockerInvocationError(t *testing.T) {
	mockClient := &MockDockerClient{
		ReturnError: io.EOF,
	}

	_, err := BuildImage(mockClient, bytes.NewReader([]byte("data")), "test-app")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "docker build failed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestBuildImage_BuildFailureOutput(t *testing.T) {
	mockClient := &MockDockerClient{
		ReturnBody: `{"error":"manifest not found"}`,
	}

	_, err := BuildImage(mockClient, bytes.NewReader([]byte("data")), "test-app")

	if err == nil {
		t.Fatal("expected build error")
	}

	if !strings.Contains(err.Error(), "manifest not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildImage_StreamsContext(t *testing.T) {
	mockClient := &MockDockerClient{}

	largeData := bytes.Repeat([]byte("test "), 10000)

	_, err := BuildImage(mockClient, bytes.NewReader(largeData), "test-app")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(mockClient.BuildContext, largeData) {
		t.Fatal("build context mismatch")
	}
}
