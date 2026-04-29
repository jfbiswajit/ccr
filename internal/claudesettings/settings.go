package claudesettings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReadSettings reads ~/.claude/settings.json as a raw map.
// Returns an empty map (not an error) if the file doesn't exist.
func ReadSettings(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("settings.json is not valid JSON: %w", err)
	}
	return settings, nil
}

// PatchStatusLine reads settings.json, updates only the statusLine field, and writes back.
// Pass nil to remove the statusLine key entirely.
func PatchStatusLine(path string, statusLine map[string]interface{}) error {
	settings, err := ReadSettings(path)
	if err != nil {
		return err
	}

	if statusLine == nil {
		delete(settings, "statusLine")
	} else {
		settings["statusLine"] = statusLine
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// validate the JSON we're about to write
	var check map[string]interface{}
	if err := json.Unmarshal(data, &check); err != nil {
		return fmt.Errorf("generated invalid JSON, aborting write: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// BackupSettings copies settings.json to backupDir/settings.YYYYMMDDHHMMSS.json.
func BackupSettings(settingsPath, backupDir string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return err
	}

	ts := time.Now().Format("20060102150405")
	dest := filepath.Join(backupDir, fmt.Sprintf("settings.%s.json", ts))
	return os.WriteFile(dest, data, 0600)
}
