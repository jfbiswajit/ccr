package cmd

import (
	"fmt"
	"strings"

	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/openrouter"
	"github.com/biswajit/ccr/internal/shell"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models [search]",
	Short: "List and select OpenRouter models",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runModels,
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}

func runModels(cmd *cobra.Command, args []string) error {
	dir := ccrDir()
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("ccr is not initialized. Run 'ccr init' first")
	}

	client := openrouter.NewClient()
	models, err := client.FetchModels(cfg.APIKey)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	query := ""
	if len(args) > 0 {
		query = args[0]
		models = openrouter.FilterModels(models, query)
	}

	if len(models) == 0 {
		fmt.Printf("No models matched %q\n", query)
		return nil
	}

	// Print table
	fmt.Printf("%-50s %-8s %-14s %-14s\n", "Model ID", "Context", "Input $/M", "Output $/M")
	fmt.Println(strings.Repeat("-", 90))
	for _, m := range models {
		id := m.ID
		if query != "" {
			id = highlight(id, query)
		}
		fmt.Printf("%-50s %-8s %-14s %-14s\n",
			id,
			formatContext(m.ContextLength),
			fmt.Sprintf("$%.4f", m.InputCost*1_000_000),
			fmt.Sprintf("$%.4f", m.OutputCost*1_000_000),
		)
	}
	fmt.Printf("\n%d model(s)\n\n", len(models))

	// Ask if user wants to set one
	var wantSelect bool
	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Set a model for a slot?").
				Value(&wantSelect),
		),
	).Run()

	if !wantSelect {
		return nil
	}

	// Pick slot
	var slot string
	huh.NewForm(
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
	).Run()

	// Pick model from the listed results
	options := make([]huh.Option[string], len(models))
	for i, m := range models {
		options[i] = huh.NewOption(m.ID, m.ID)
	}

	var modelID string
	huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Model").
				Options(options...).
				Value(&modelID),
		),
	).Run()

	if slot == "" || modelID == "" {
		return nil
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

func formatContext(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

func highlight(s, query string) string {
	lower := strings.ToLower(s)
	q := strings.ToLower(query)
	idx := strings.Index(lower, q)
	if idx < 0 {
		return s
	}
	return s[:idx] + "\033[1m" + s[idx:idx+len(query)] + "\033[0m" + s[idx+len(query):]
}
