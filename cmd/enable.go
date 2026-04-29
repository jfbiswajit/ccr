package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/biswajit/ccr/internal/claudesettings"
	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/shell"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Switch Claude Code to OpenRouter",
	RunE:  runEnable,
}

func init() {
	rootCmd.AddCommand(enableCmd)
}

func runEnable(cmd *cobra.Command, args []string) error {
	dir := ccrDir()
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("ccr is not initialized. Run 'ccr init' first")
	}

	if cfg.ActiveProvider == "openrouter" {
		fmt.Println("Already enabled.")
		return nil
	}

	settingsPath := claudeSettingsPath()
	backupDir := filepath.Join(dir, "backups")

	// Backup settings.json before first patch
	if err := claudesettings.BackupSettings(settingsPath, backupDir); err != nil {
		// Non-fatal if settings.json doesn't exist yet
		_ = err
	}

	// Rewrite env file with OpenRouter vars
	if err := shell.WriteEnvFile(envFilePath(), true, cfg.APIKey, cfg.ModelOverrides); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	// Patch statusLine in settings.json
	statusLine := map[string]interface{}{
		"type":    "command",
		"command": fmt.Sprintf("sh %s", filepath.Join(dir, "statusline.sh")),
	}
	if err := claudesettings.PatchStatusLine(settingsPath, statusLine); err != nil {
		return fmt.Errorf("failed to update ~/.claude/settings.json: %w", err)
	}

	// Update state
	if err := config.SaveConfig(dir, &config.Config{ActiveProvider: "openrouter"}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("OpenRouter enabled. Run: source %s\n", envFilePath())
	return nil
}
