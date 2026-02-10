package shared

const (
    UploadEndpoint = "/upload"
    UploadPort     = 8080
    AppPort        = 8888
    ContentTypeTGZ = "application/x-gzip"
)

type UploadResponse struct {
    URL   string `json:"url,omitempty"`
    Error string `json:"error,omitempty"`
}
