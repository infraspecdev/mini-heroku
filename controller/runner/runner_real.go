package runner

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type RealRunnerClient struct {
	client *client.Client
}

func NewRealRunnerClient() (*RealRunnerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &RealRunnerClient{client: cli}, nil
}

func (r *RealRunnerClient) ContainerCreate(ctx context.Context, config ContainerConfig, hostConfig HostConfig) (ContainerCreateResponse, error) {
	// Convert exposed ports
	exposedPorts := nat.PortSet{}
	for port := range config.ExposedPorts {
		natPort, err := nat.NewPort("tcp", port[:len(port)-4]) // Remove "/tcp"
		if err != nil {
			return ContainerCreateResponse{}, err
		}
		exposedPorts[natPort] = struct{}{}
	}

	// Convert port bindings
	portBindings := nat.PortMap{}
	for port, bindings := range hostConfig.PortBindings {
		natPort, err := nat.NewPort("tcp", port[:len(port)-4]) // Remove "/tcp"
		if err != nil {
			return ContainerCreateResponse{}, err
		}
		portBindingList := []nat.PortBinding{}
		for _, binding := range bindings {
			portBindingList = append(portBindingList, nat.PortBinding{
				HostIP:   binding.HostIP,
				HostPort: binding.HostPort,
			})
		}
		portBindings[natPort] = portBindingList
	}

	// Create container
	containerConfig := &container.Config{
		Image:        config.Image,
		ExposedPorts: exposedPorts,
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
	}

	resp, err := r.client.ContainerCreate(ctx, containerConfig, hostCfg, nil, nil, "")
	if err != nil {
		return ContainerCreateResponse{}, err
	}

	return ContainerCreateResponse{ID: resp.ID}, nil
}

func (r *RealRunnerClient) ContainerStart(ctx context.Context, containerID string) error {
	return r.client.ContainerStart(ctx, containerID, container.StartOptions{})
}
