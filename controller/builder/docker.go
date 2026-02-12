package builder

import (
	"context"
	"fmt"
	"io"
)

type ImageBuildOptions struct {
	Tags       []string
	DockerFile string
	Remove     bool
}
type ImageBuildResponse struct {
	Body io.ReadCloser
}

type DockerClient interface {
	ImageBuild(ctx context.Context, buildContext io.Reader, options ImageBuildOptions)(ImageBuildResponse, error)
}

func BuildImage(client DockerClient, tarballReader io.Reader, appName string) (string, error){
	options := ImageBuildOptions{
		Tags: []string{appName + ":latest"},
		DockerFile: "Dockefile",
		Remove: true,

	}

	resp, err := client.ImageBuild(context.Background(), tarballReader, options)

	if err != nil{
		return "", fmt.Errorf("docker build failed: %w",err)

	}

	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	return appName + ":latest", nil
}