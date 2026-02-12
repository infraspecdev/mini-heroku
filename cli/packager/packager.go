package packager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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


// -----------------------------
// CreateTarball
// Creates in-memory tar.gz archive
// -----------------------------

func CreateTarball(files map[string]string) ([]byte, error) {
	var buf bytes.Buffer

	// gzip wraps tar
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for name, content := range files {

		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("writing header: %w", err)
		}

		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("writing content: %w", err)
		}
	}

	// Close in reverse order
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("closing tar writer: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("closing gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}