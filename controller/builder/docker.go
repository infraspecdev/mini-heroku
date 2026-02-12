package builder

import (
	"bufio"
	"context"
	"encoding/json"
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
	ImageBuild(ctx context.Context, buildContext io.Reader, options ImageBuildOptions) (ImageBuildResponse, error)
}

type buildMessage struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

func BuildImage(client DockerClient, tarballReader io.Reader, appName string) (string, error) {
	options := ImageBuildOptions{
		Tags:       []string{appName + ":latest"},
		DockerFile: "Dockefile",
		Remove:     true,
	}

	resp, err := client.ImageBuild(context.Background(), tarballReader, options)

	if err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)

	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var buildError string

	for scanner.Scan() {
		var msg buildMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		if msg.Error != "" {
			buildError = msg.Error
		}
	}

	if buildError != "" {
		return "", fmt.Errorf("docker build error: %s", buildError)
	}

	return appName + ":latest", nil
}