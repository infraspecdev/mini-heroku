package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"mini-heroku/cli/config"

	"github.com/spf13/cobra"
)

// logsCmd is registered in init() below and wired into rootCmd.
var logsCmd = &cobra.Command{
	Use:   "logs <appName>",
	Short: "Stream live logs from a deployed app",
	Long: `Stream stdout and stderr from the Docker container running <appName>.

Behaves like 'docker logs -f': output is printed in real-time and the command
blocks until the container exits or you press Ctrl-C.`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

func runLogs(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	host := cfg.ServerURL
	if host == "" {
		host = "http://localhost:8080"
	}

	url := fmt.Sprintf("%s/apps/%s/logs", host, appName)

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	httpClient := &http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connecting to controller: %w", err)
	}
	defer resp.Body.Close()

	// Surface controller-side errors cleanly.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("controller returned %d: %s", resp.StatusCode, string(body))
	}

	fmt.Fprintf(os.Stderr, "=== logs for %s (Ctrl-C to stop) ===\n", appName)

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		// io.EOF and context-cancellation errors are both normal exit paths.
		if err != io.EOF {
			return fmt.Errorf("log stream interrupted: %w", err)
		}
	}

	return nil
}
