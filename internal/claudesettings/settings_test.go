package claudesettings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, _ := json.MarshalIndent(v, "", "  ")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, data, 0644)
}

func TestReadSettingsMissingFile(t *testing.T) {
	dir := t.TempDir()
	s, err := ReadSettings(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
}

func TestPatchOnlyChangesStatusLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	original := map[string]interface{}{
		"effortLevel":  "high",
		"voiceEnabled": true,
		"hooks":        map[string]interface{}{"Stop": []string{"afplay glass.aiff"}},
		"statusLine":   map[string]interface{}{"type": "command", "command": "sh ~/.claude/statusline.sh"},
	}
	writeJSON(t, path, original)

	if err := PatchStatusLine(path, map[string]interface{}{"type": "command", "command": "sh ~/.ccr/statusline.sh"}); err != nil {
		t.Fatalf("patch error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["effortLevel"] != "high" {
		t.Errorf("effortLevel was modified: %v", result["effortLevel"])
	}
	if result["voiceEnabled"] != true {
		t.Errorf("voiceEnabled was modified: %v", result["voiceEnabled"])
	}
	sl := result["statusLine"].(map[string]interface{})
	if sl["command"] != "sh ~/.ccr/statusline.sh" {
		t.Errorf("statusLine not updated: %v", sl["command"])
	}
}

func TestPatchRemovesStatusLineWhenNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	original := map[string]interface{}{
		"effortLevel": "high",
		"statusLine":  map[string]interface{}{"type": "command", "command": "old"},
	}
	writeJSON(t, path, original)

	if err := PatchStatusLine(path, nil); err != nil {
		t.Fatalf("patch error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if _, exists := result["statusLine"]; exists {
		t.Error("expected statusLine to be removed")
	}
	if result["effortLevel"] != "high" {
		t.Errorf("effortLevel was modified: %v", result["effortLevel"])
	}
}

func TestBackupCreatesTimestampedFile(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	backupDir := filepath.Join(dir, "backups")

	writeJSON(t, settingsPath, map[string]interface{}{"key": "value"})

	if err := BackupSettings(settingsPath, backupDir); err != nil {
		t.Fatalf("backup error: %v", err)
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("readdir error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 backup file, got %d", len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "settings.") {
		t.Errorf("unexpected backup filename: %s", entries[0].Name())
	}
}

func TestReadStatusLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	writeJSON(t, path, map[string]interface{}{
		"statusLine": map[string]interface{}{"type": "command", "command": "sh old.sh"},
	})

	s, err := ReadSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s["statusLine"] == nil {
		t.Error("expected statusLine to be present")
	}
}
