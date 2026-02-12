package builder

import "io"

type ImageBuildOptions struct {
	Tags       []string
	DockerFile string
	Remove     bool
}
type ImageBuildResponse struct {
	Body io.ReadCloser
}