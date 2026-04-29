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

	// Step 1: pick which slot to configure
	var slot string
	slotForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which model slot?").
				Description("Claude Code uses different models for different task types").
				Options(
					huh.NewOption("Sonnet  (general coding)", "sonnet"),
					huh.NewOption("Opus    (complex reasoning)", "opus"),
					huh.NewOption("Haiku   (quick completions)", "haiku"),
					huh.NewOption("Subagent (sub-agent tasks)", "subagent"),
				).
				Value(&slot),
		),
	)
	if err := slotForm.Run(); err != nil {
		return err
	}

	// Step 2: fetch models
	fmt.Print("Fetching models... ")
	client := openrouter.NewClient()
	models, err := client.FetchModels(cfg.APIKey)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}
	fmt.Printf("%d available\n", len(models))

	// Build options for picker
	options := make([]huh.Option[string], len(models))
	for i, m := range models {
		label := fmt.Sprintf("%-50s  $%.4f / $%.4f per 1M tokens", m.ID, m.InputCost*1_000_000, m.OutputCost*1_000_000)
		options[i] = huh.NewOption(label, m.ID)
	}

	// Step 3: pick the model
	var modelID string
	modelForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Pick a model for %s", slot)).
				Options(options...).
				Value(&modelID),
		),
	)
	if err := modelForm.Run(); err != nil {
		return err
	}

	// Save to config
	overrides := cfg.ModelOverrides
	if overrides == nil {
		overrides = map[string]string{}
	}
	overrides[slot] = modelID

	if err := config.SaveConfig(dir, &config.Config{ModelOverrides: overrides}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Rewrite env.sh with updated overrides
	enabled := cfg.ActiveProvider == "openrouter"
	if err := shell.WriteEnvFile(envFilePath(), enabled, cfg.APIKey, overrides); err != nil {
		return fmt.Errorf("failed to update env file: %w", err)
	}

	fmt.Printf("\nSet %s → %s\n", slot, modelID)
	fmt.Printf("Run: source %s\n", envFilePath())
	return nil
}
