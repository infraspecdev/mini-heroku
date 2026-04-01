package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"mini-heroku/cli/config"
	"mini-heroku/cli/keychain"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var setAPIKeyCmd = &cobra.Command{
	Use:   "set-api-key <key>",
	Short: "Save your API key to the OS keychain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.TrimSpace(args[0])
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}
		if err := keychain.Set(key); err != nil {
			return fmt.Errorf("storing API key: %w", err)
		}
		fmt.Println("API key saved to OS keychain")
		return nil
	},
}

var setHostCmd = &cobra.Command{
	Use:   "set-host <url>",
	Short: "Set the controller server URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
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
	configCmd.AddCommand(setHostCmd, getHostCmd, setAPIKeyCmd)
}
