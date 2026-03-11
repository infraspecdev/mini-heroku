package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mini version %s\n", Version)
	},
}
