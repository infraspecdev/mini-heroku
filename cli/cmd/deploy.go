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

	Run: func(cmd *cobra.Command, args []string) {
		folder := args[0]
		appName := args[1]

		cfg, err:=config.Load()
		if err != nil{
			fmt.Println("Failed to load config:", err)
			return
		}

		host := cfg.ServerURL
		if host == ""{
			fmt.Println("Controller host not configured")
			fmt.Println("Run: mini config set-host <url>")
			return
		}


		fmt.Println("Discovering files...")

		files, err := packager.ExploreDirectory(folder)
		if err != nil{
			fmt.Println("Failed to explore directory:", err)
			return
		}

		fmt.Printf("Found %d files\n", len(files))

		fileMap := make(map[string]string)

		for _,f := range files {
			fullPath := filepath.Join(folder,f)

			data, err := os.ReadFile(fullPath)
			if err != nil {
				fmt.Println("Failed to read file:", f)
				return
			}

			fileMap[f]=string(data)
		}

		fmt.Println("Creating archive...")

		tarball, err := packager.CreateTarball(fileMap)
		if err != nil {
			fmt.Println("Failed to create tarball:", err)
			return
		}

		fmt.Printf("Archive size: %d bytes\n", len(tarball))

		fmt.Printf("Uploading to server...")

		reader := bytes.NewReader(tarball)

		resp, err := client.UploadPackage(host, reader, appName)
		if err != nil{
			fmt.Println("Deployment failed:", err)
			return
		}

		fmt.Println("")
		fmt.Println("Status :", resp.Status)
		fmt.Println("Message:", resp.Message)

		if resp.AppURL != "" {
			fmt.Println("App URL:", resp.AppURL)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

}
