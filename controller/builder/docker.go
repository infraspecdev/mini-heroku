package builder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	DefaultDockerfile = "Dockerfile"
	ImageTagSuffix    = ":latest"
)

type ImageBuildOptions struct {
	Tags       []string
	Dockerfile string
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

func NewImageBuildOptions(appName string) ImageBuildOptions {
	return ImageBuildOptions{
		Tags:       []string{appName + ImageTagSuffix},
		Dockerfile: DefaultDockerfile,
		Remove:     true,
	}
}

func BuildImage(client DockerClient, tarballReader io.Reader, appName string) (string, error) {
	options := NewImageBuildOptions(appName)

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
			continue //skip malformed lines
		}

		// Check for errors in build output
		if msg.Error != "" {
			buildError = msg.Error
		}

		// Log build progress
		if msg.Stream != "" {
            msg.Stream = strings.TrimSpace(msg.Stream)
            if msg.Stream != "" {
                fmt.Printf("Build: %s\n", msg.Stream)
            }
        }
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading build output: %w", err)
	}

	if buildError != "" {
		return "", fmt.Errorf("docker build error: %s", buildError)
	}

	return appName + ":latest", nil
}
