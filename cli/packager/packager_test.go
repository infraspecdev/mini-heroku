package packager

import (
    "os"
    "path/filepath"
    "testing"
	"archive/tar"
    "bytes"
    "compress/gzip"
    "io"

)

func TestFileDiscovery(t *testing.T) {
    // Setup: Create temp directory with test files
    tmpDir := t.TempDir()
    
    // Create valid files
    createFile(t, tmpDir, "main.go", "package main")
    createFile(t, tmpDir, "handler.go", "package handler")
    
    // Create files to ignore
    createFile(t, tmpDir, ".git/config", "gitconfig")
    createFile(t, tmpDir, ".env", "SECRET=key")
    createFile(t, tmpDir, "node_modules/lib.js", "module")
    
    // Execute
    files, err := ExploreDirectory(tmpDir)
    
    // Assert
    if err != nil {
        t.Fatalf("ExploreDirectory failed: %v", err)
    }
    
    // Should return 2 files
    if len(files) != 2 {
        t.Errorf("Expected 2 files, got %d", len(files))
    }
    
    // Should contain main.go and handler.go
    expected := map[string]bool{
        "main.go": true,
        "handler.go": true,
    }
    
    for _, file := range files {
        if !expected[file] {
            t.Errorf("Unexpected file: %s", file)
        }
    }
    
    // Should NOT contain .git, .env, or node_modules
    for _, file := range files {
        if filepath.Dir(file) == ".git" {
            t.Error("Should exclude .git directory")
        }
    }
}

func createFile(t *testing.T, dir, path, content string) {
    fullPath := filepath.Join(dir, path)
    os.MkdirAll(filepath.Dir(fullPath), 0755)
    if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
        t.Fatal(err)
    }
}


func TestTarballGeneration(t *testing.T) {
    // Setup: Define test files
    files := map[string]string{
        "main.go": "package main\n\nfunc main() {}",
        "handler.go": "package handler",
        "README.md": "# My App",
    }
    
    // Execute: Create tarball
    tarballBytes, err := CreateTarball(files)
    if err != nil {
        t.Fatalf("CreateTarball failed: %v", err)
    }
    
    // Assert: Verify it's a valid tar.gz
    // Step 1: Decompress gzip
    gzipReader, err := gzip.NewReader(bytes.NewReader(tarballBytes))
    if err != nil {
        t.Fatalf("Invalid gzip: %v", err)
    }
    defer gzipReader.Close()
    
    // Step 2: Read tar contents
    tarReader := tar.NewReader(gzipReader)
    
    foundFiles := make(map[string]string)
    
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break // End of tar archive
        }
        if err != nil {
            t.Fatalf("Error reading tar: %v", err)
        }
        
        // Read file content
        content := make([]byte, header.Size)
        if _, err := io.ReadFull(tarReader, content); err != nil {
            t.Fatalf("Error reading file content: %v", err)
        }
        
        foundFiles[header.Name] = string(content)
    }
    
    // Assert: All files present with correct content
    if len(foundFiles) != len(files) {
        t.Errorf("Expected %d files, got %d", len(files), len(foundFiles))
    }
    
    for name, expectedContent := range files {
        actualContent, exists := foundFiles[name]
        if !exists {
            t.Errorf("File %s missing from tarball", name)
            continue
        }
        if actualContent != expectedContent {
            t.Errorf("File %s content mismatch.\nExpected: %q\nGot: %q", 
                name, expectedContent, actualContent)
        }
    }
}

