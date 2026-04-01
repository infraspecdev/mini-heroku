package config_test

import (
	"mini-heroku/cli/config"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadServerURL(t *testing.T) {
	// Use temp dir to avoid touching real ~/.mini
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &config.Config{
		ServerURL: "https://example.test",
	}

	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ServerURL != cfg.ServerURL {
		t.Errorf("expected ServerURL %q, got %q", cfg.ServerURL, loaded.ServerURL)
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load should not fail when file missing: %v", err)
	}
	if cfg.ServerURL != "" {
		t.Error("expected empty ServerURL on fresh config")
	}
}

func TestSave_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &config.Config{ServerURL: "https://example.test"}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	configFile := filepath.Join(tmp, ".mini", "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}
