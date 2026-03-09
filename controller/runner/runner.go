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

type ContainerInspectResponse struct {
	IPAddress string
}

type RunResult struct {
	ContainerID string
	ContainerIP string
	HostPort    string
}

type RunnerClient interface {
	ContainerCreate(ctx context.Context, config ContainerConfig, hostConfig HostConfig) (ContainerCreateResponse, error)
	ContainerStart(ctx context.Context, containerID string) error
	ContainerInspect(ctx context.Context, containerID string) (ContainerInspectResponse, error)
}

func RunContainer(client RunnerClient, imageName string, hostPort int) (*RunResult, error) {
	ctx := context.Background()

	exposedPorts := map[string]struct{}{
		"8080/tcp": {},
	}

	portBindings := map[string][]PortBinding{
		"8080/tcp": {
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", hostPort),
			},
		},
	}

	config := ContainerConfig{
		Image:        imageName,
		ExposedPorts: exposedPorts,
	}

	hostConfig := HostConfig{
		PortBindings: portBindings,
	}

	resp, err := client.ContainerCreate(ctx, config, hostConfig)
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}

	if err := client.ContainerStart(ctx, resp.ID); err != nil {
		return nil, fmt.Errorf("starting container: %w", err)
	}

	inspect, err := client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	return &RunResult{
		ContainerID: resp.ID,
		ContainerIP: inspect.IPAddress,
		HostPort:    fmt.Sprintf("%d", hostPort),
	}, nil

}

func GenerateHostPort(appName string) int {
	hash := 0
	for _, c := range appName {
		hash += int(c)
	}

	port := 10000 + (hash % 10000)
	return port
}
