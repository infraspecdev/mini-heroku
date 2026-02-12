package runner

import (
	"context"
	"testing"
)

type MockRunnerClient struct {
	CreateCalled     bool
	CreateConfig     ContainerConfig
	CreateHostConfig HostConfig
	StartCalled      bool
	ContainerID      string
}

func (m *MockRunnerClient) ContainerCreate(ctx context.Context, config ContainerConfig, hostConfig HostConfig) (ContainerCreateResponse, error) {
	m.CreateCalled = true
	m.CreateConfig = config
	m.CreateHostConfig = hostConfig
	m.ContainerID = "container-abc123"

	return ContainerCreateResponse{ID: m.ContainerID}, nil
}

func (m *MockRunnerClient) ContainerStart(ctx context.Context, containerID string) error {
	m.StartCalled = true
	return nil
}

func TestRunContainer_Success(t *testing.T) {
	mockClient := &MockRunnerClient{}

	containerURL, err := RunContainer(mockClient, "my-app:latest", 8888)

	if err != nil {
		t.Fatalf("RunContainer failed: %v", err)
	}

	if !mockClient.CreateCalled {
		t.Error("ContainerCreate was not called")
	}

	if !mockClient.StartCalled {
		t.Error("ContainerStart was not called")
	}

	expectedURL := "http://localhost:8888"
	if containerURL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, containerURL)
	}
}
