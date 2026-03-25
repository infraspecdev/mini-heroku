package cmd

import (
	"bytes"
	"fmt"
	"mini-heroku/cli/client"
	"mini-heroku/cli/config"
	"mini-heroku/cli/packager"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [folder] [app-name]",
	Short: "Deploy an application to the mini platform",
	Args: cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		folder := args[0]
		appName := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		serverURL := cfg.ServerURL
		if serverURL == "" {
			fmt.Println("Controller host not configured")
			fmt.Println("Run: mini config set-host <url>")
			return nil
		}

		if cfg.APIKey == "" {
			return fmt.Errorf("no API key configured — run: mini config set-api-key <key>")
		}

		fmt.Println("Discovering files...")

		files, err := packager.ExploreDirectory(folder)
		if err != nil {
			return fmt.Errorf("exploring directory: %w", err)
		}

		fmt.Printf("Found %d files\n", len(files))

		fileMap := make(map[string]string)

		for _, f := range files {
			fullPath := filepath.Join(folder, f)

			data, err := os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", f, err)
			}

			fileMap[f] = string(data)
		}

		fmt.Println("Creating archive...")

		tarball, err := packager.CreateTarball(fileMap)
		if err != nil {
			return fmt.Errorf("creating tarball: %w", err)
		}

		fmt.Printf("Archive size: %d bytes\n", len(tarball))

		fmt.Printf("Uploading to server...")

		reader := bytes.NewReader(tarball)

		resp, err := client.UploadPackage(serverURL, reader, appName, cfg.APIKey)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		fmt.Println("")
		fmt.Println("Status :", resp.Status)
		fmt.Println("Message:", resp.Message)

		if resp.AppURL != "" {
			fmt.Println("App URL:", resp.AppURL)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

}
