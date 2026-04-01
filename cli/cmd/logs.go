package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"mini-heroku/cli/client"
	"mini-heroku/cli/config"
	"mini-heroku/cli/keychain"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <appName>",
	Short: "Stream live logs from a deployed app",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogs,
}

func runLogs(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	apiKey, err := keychain.Get()
	if err != nil {
		return err
	}

	host := cfg.ServerURL
	if host == "" {
		host = "http://localhost:8080"
	}

	url := fmt.Sprintf("%s/apps/%s/logs", host, appName)

	// Use signed request so the controller can verify HMAC + timestamp.
	req, err := client.NewSignedRequest(http.MethodGet, url, apiKey, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req = req.WithContext(cmd.Context())

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("connecting to controller: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return fmt.Errorf("controller returned %d: %s", resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("controller returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Fprintf(os.Stderr, "=== logs for %s (Ctrl-C to stop) ===\n", appName)
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil && err != io.EOF {
		return fmt.Errorf("log stream interrupted: %w", err)
	}
	return nil
}
