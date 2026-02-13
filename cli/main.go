package main

import (
	"bytes"
	"fmt"
	"mini-heroku/cli/client"
	"mini-heroku/cli/packager"
	"os"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "deploy" {
	fmt.Println("Usage: mini deploy <directory>")
	os.Exit(1)
	}

	appDir := os.Args[2]

	// 1. Discover files
	fmt.Println("📦 Discovering files...")
	files, err := packager.ExploreDirectory(appDir)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("❌ No files found in directory")
		os.Exit(1)
	}

	fmt.Printf("   Found %d files\n", len(files))

	// 2. Read file contents
	fileContents := make(map[string]string)
	for _, file := range files {
		fullPath := appDir + string(os.PathSeparator) + file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("❌ Error reading %s: %v\n", file, err)
			os.Exit(1)
		}
		fileContents[file] = string(content)
	}

	// 3. Create tarball
	fmt.Println("🗜️  Creating archive...")
	tarballBytes, err := packager.CreateTarball(fileContents)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   Archive size: %d bytes\n", len(tarballBytes))

	// 4. Upload
	fmt.Println("🚀 Uploading to server...")
	response, err := client.UploadPackage(
		client.DefaultServerURL,
		bytes.NewReader(tarballBytes),
		"my-app",
	)

	if err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
		os.Exit(1)
	}

	// 5. Display result
	if response.Status == client.StatusSuccess {
		fmt.Printf("✅ Success! Your app is live at: %s\n", response.AppURL)
	} else {
		fmt.Printf("❌ Deployment failed: %s\n", response.Message)
		os.Exit(1)
	}
}
