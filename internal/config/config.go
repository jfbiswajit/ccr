package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey             string                 `json:"apiKey"`
	ActiveProvider     string                 `json:"activeProvider"`
	OriginalStatusLine map[string]interface{} `json:"originalStatusLine,omitempty"`
}

func configFilePath(dir string) string {
	return filepath.Join(dir, "config.json")
}

func LoadConfig(dir string) (*Config, error) {
	path := configFilePath(dir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{ActiveProvider: "anthropic"}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.ActiveProvider == "" {
		cfg.ActiveProvider = "anthropic"
	}
	return &cfg, nil
}

func SaveConfig(dir string, cfg *Config) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// merge: load existing then overlay new values
	existing, err := LoadConfig(dir)
	if err != nil {
		existing = &Config{ActiveProvider: "anthropic"}
	}
	if cfg.APIKey != "" {
		existing.APIKey = cfg.APIKey
	}
	if cfg.ActiveProvider != "" {
		existing.ActiveProvider = cfg.ActiveProvider
	}
	if cfg.OriginalStatusLine != nil {
		existing.OriginalStatusLine = cfg.OriginalStatusLine
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	// atomic write: write to tmp then rename
	tmp := configFilePath(dir) + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, configFilePath(dir))
}
