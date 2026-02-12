package builder

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type RealDockerClient struct {
	client *client.Client
}

func NewRealDockerClient() (*RealDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &RealDockerClient{client: cli}, nil
}

func (r *RealDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options ImageBuildOptions) (ImageBuildResponse, error) {
	dockerOptions := types.ImageBuildOptions{
		Tags:       options.Tags,
		Dockerfile: options.Dockerfile,
		Remove:     options.Remove,
	}

	response, err := r.client.ImageBuild(ctx, buildContext, dockerOptions)
	if err != nil {
		return ImageBuildResponse{}, err
	}

	return ImageBuildResponse{Body: response.Body}, nil
}