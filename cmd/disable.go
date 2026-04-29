package cmd

import (
	"fmt"

	"github.com/biswajit/ccr/internal/claudesettings"
	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/shell"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Switch Claude Code back to Anthropic",
	RunE:  runDisable,
}

func init() {
	rootCmd.AddCommand(disableCmd)
}

func runDisable(cmd *cobra.Command, args []string) error {
	dir := ccrDir()
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("ccr is not initialized. Run 'ccr init' first")
	}

	if cfg.ActiveProvider == "anthropic" {
		fmt.Println("Already disabled.")
		return nil
	}

	// Rewrite env file to neutral (unsets only)
	if err := shell.WriteEnvFile(envFilePath(), false, "", nil); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	// Restore original statusLine in settings.json
	if err := claudesettings.PatchStatusLine(claudeSettingsPath(), cfg.OriginalStatusLine); err != nil {
		return fmt.Errorf("failed to restore ~/.claude/settings.json: %w", err)
	}

	// Update state
	if err := config.SaveConfig(dir, &config.Config{ActiveProvider: "anthropic"}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Anthropic restored. Run: source %s\n", envFilePath())
	return nil
}
