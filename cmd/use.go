package cmd

import (
	"fmt"

	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/openrouter"
	"github.com/biswajit/ccr/internal/shell"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Pick a model for Claude Code to use via OpenRouter",
	RunE:  runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func runUse(cmd *cobra.Command, args []string) error {
	dir := ccrDir()
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("ccr is not initialized. Run 'ccr init' first")
	}

	// Fetch models silently before showing any prompts
	client := openrouter.NewClient()
	models, err := client.FetchModels(cfg.APIKey)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	// Build model options
	options := make([]huh.Option[string], len(models))
	for i, m := range models {
		options[i] = huh.NewOption(m.ID, m.ID)
	}

	var slot, modelID string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Slot").
				Options(
					huh.NewOption("sonnet", "sonnet"),
					huh.NewOption("opus", "opus"),
					huh.NewOption("haiku", "haiku"),
					huh.NewOption("subagent", "subagent"),
				).
				Value(&slot),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Model").
				Options(options...).
				Value(&modelID),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	overrides := cfg.ModelOverrides
	if overrides == nil {
		overrides = map[string]string{}
	}
	overrides[slot] = modelID

	if err := config.SaveConfig(dir, &config.Config{ModelOverrides: overrides}); err != nil {
		return err
	}

	enabled := cfg.ActiveProvider == "openrouter"
	if err := shell.WriteEnvFile(envFilePath(), enabled, cfg.APIKey, overrides); err != nil {
		return err
	}

	fmt.Printf("Done. source %s\n", envFilePath())
	return nil
}
