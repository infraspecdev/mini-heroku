package packager

import (
    "os"
    "path/filepath"
    "testing"
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