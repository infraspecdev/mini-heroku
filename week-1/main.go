package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func listContainers(ctx context.Context, cli *client.Client) {
	result, err := cli.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nRunning Containers:")
	for _, c := range result.Items {
		fmt.Printf("ID: %s | Image: %s | Status: %s\n",
			c.ID[:12], c.Image, c.Status)
	}
}


func pullAndRunAlpine(ctx context.Context, cli *client.Client) {
	fmt.Println("\nPulling alpine image...")

	result, err := cli.ImagePull(ctx, "docker.io/library/alpine", client.ImagePullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, result)

	resp, err := cli.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Config: &container.Config{
				Image: "alpine",
				Cmd:   []string{"echo", "Hello from Mini PaaS"},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nStarted Alpine container:", resp.ID[:12])
}


func main() {
	ctx := context.Background()

	// Connect to Docker daemon
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Print Docker version
	version, err := cli.ServerVersion(ctx, client.ServerVersionOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Docker Version:", version.Version)
	fmt.Println("Docker API Version:", version.APIVersion)

	listContainers(ctx, cli)

	pullAndRunAlpine(ctx, cli)


}
