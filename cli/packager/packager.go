package packager

import (
	"fmt"
	"os"
	"path/filepath"
)


// -----------------------------
// Ignore Patterns
// -----------------------------

var ignoredDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
	"__pycache__": true,
	".vscode":     true,
	".idea":       true,
}

var ignoredFiles = map[string]bool{
	".env": true,
}


// -----------------------------
// ExploreDirectory
// Recursively walks a directory
// and returns valid relative file paths
// -----------------------------

func ExploreDirectory(rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip root itself
		if path == rootPath {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}

		// Normalize to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Check directory ignore
		if info.IsDir() {
			dirName := info.Name()
			if ignoredDirs[dirName] {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include regular files
		if !info.Mode().IsRegular() {
			return nil
		}

		// Ignore specific files
		if ignoredFiles[info.Name()] {
			return nil
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}