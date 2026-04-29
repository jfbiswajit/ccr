package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ActiveProvider != "anthropic" {
		t.Errorf("expected default activeProvider 'anthropic', got %q", cfg.ActiveProvider)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty apiKey, got %q", cfg.APIKey)
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		APIKey:         "test-key-123",
		ActiveProvider: "openrouter",
	}
	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("apiKey mismatch: got %q want %q", loaded.APIKey, cfg.APIKey)
	}
	if loaded.ActiveProvider != cfg.ActiveProvider {
		t.Errorf("activeProvider mismatch: got %q want %q", loaded.ActiveProvider, cfg.ActiveProvider)
	}
}

func TestSaveCreatesDirectoryWithCorrectPerms(t *testing.T) {
	dir := t.TempDir()
	ccrDir := filepath.Join(dir, ".ccr")

	if err := SaveConfig(ccrDir, &Config{APIKey: "k", ActiveProvider: "anthropic"}); err != nil {
		t.Fatalf("save error: %v", err)
	}

	info, err := os.Stat(ccrDir)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .ccr to be a directory")
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("expected dir perm 0700, got %o", info.Mode().Perm())
	}

	cfgFile := filepath.Join(ccrDir, "config.json")
	finfo, err := os.Stat(cfgFile)
	if err != nil {
		t.Fatalf("config file stat error: %v", err)
	}
	if finfo.Mode().Perm() != 0600 {
		t.Errorf("expected file perm 0600, got %o", finfo.Mode().Perm())
	}
}

func TestSaveDoesNotWipeUnrelatedFields(t *testing.T) {
	dir := t.TempDir()

	first := &Config{APIKey: "key-1", ActiveProvider: "anthropic", OriginalStatusLine: map[string]interface{}{"type": "command"}}
	if err := SaveConfig(dir, first); err != nil {
		t.Fatalf("first save error: %v", err)
	}

	second := &Config{APIKey: "key-2", ActiveProvider: "openrouter"}
	if err := SaveConfig(dir, second); err != nil {
		t.Fatalf("second save error: %v", err)
	}

	loaded, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.APIKey != "key-2" {
		t.Errorf("expected apiKey 'key-2', got %q", loaded.APIKey)
	}
	if loaded.OriginalStatusLine == nil {
		t.Error("expected originalStatusLine to survive second save")
	}
}
