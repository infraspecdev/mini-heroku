package runner

import (
	"context"
	"fmt"
)

type ContainerConfig struct {
	Image        string
	ExposedPorts map[string]struct{}
}

type HostConfig struct {
	PortBindings map[string][]PortBinding
}

type PortBinding struct {
	HostIP   string
	HostPort string
}

type ContainerCreateResponse struct {
	ID string
}

type RunnerClient interface {
	ContainerCreate(ctx context.Context, config ContainerConfig, hostConfig HostConfig) (ContainerCreateResponse, error)
	ContainerStart(ctx context.Context, containerID string) error
}

func RunContainer(client RunnerClient, imageName string, hostPort int) (string, error) {
	ctx := context.Background()

	config := ContainerConfig{
		Image: imageName,
	}

	hostConfig := HostConfig{}

	resp, err := client.ContainerCreate(ctx, config, hostConfig)
	if err != nil {
		return "", err
	}

	if err := client.ContainerStart(ctx, resp.ID); err != nil {
		return "", err
	}

	return fmt.Sprintf("http://localhost:%d", hostPort), nil
}
