package cmd

import (
	"fmt"
	"net/url"

	"mini-heroku/cli/config"

	"github.com/spf13/cobra"
)

// configCmd is the parent: `mini config`
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

// setHostCmd: `mini config set-host <url>`
var setHostCmd = &cobra.Command{
	Use:   "set-host <url>",
	Short: "Set the controller server URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]

		// Validate the URL before saving
		if _, err := url.ParseRequestURI(rawURL); err != nil {
			return fmt.Errorf("invalid URL %q: %w", rawURL, err)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.ServerURL = rawURL

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Host set to: %s\n", rawURL)
		return nil
	},
}

// getHostCmd: `mini config get-host`
var getHostCmd = &cobra.Command{
	Use:   "get-host",
	Short: "Print the current server URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.ServerURL == "" {
			fmt.Println("No host configured. Run: mini config set-host <url>")
			return nil
		}

		fmt.Println(cfg.ServerURL)
		return nil
	},
}

func init() {
	configCmd.AddCommand(setHostCmd)
	configCmd.AddCommand(getHostCmd)
}
