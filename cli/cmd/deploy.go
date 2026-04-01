package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"mini-heroku/cli/client"
	"mini-heroku/cli/config"
	"mini-heroku/cli/keychain"
	"mini-heroku/cli/packager"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [folder] [app-name]",
	Short: "Deploy an application to the mini platform",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		folder, appName := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if cfg.ServerURL == "" {
			fmt.Println("Controller host not configured")
			fmt.Println("Run: mini config set-host <url>")
			return nil
		}

		apiKey, err := keychain.Get()
		if err != nil {
			return err
		}

		fmt.Println("Discovering files...")
		files, err := packager.ExploreDirectory(folder)
		if err != nil {
			return fmt.Errorf("exploring directory: %w", err)
		}
		fmt.Printf("Found %d files\n", len(files))

		fileMap := make(map[string]string)
		for _, f := range files {
			data, err := os.ReadFile(filepath.Join(folder, f))
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
		resp, err := client.UploadPackage(cfg.ServerURL, bytes.NewReader(tarball), appName, apiKey)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		fmt.Println()
		fmt.Println("Status :", resp.Status)
		fmt.Println("Message:", resp.Message)
		if resp.AppURL != "" {
			fmt.Println("App URL:", resp.AppURL)
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(deployCmd) }
