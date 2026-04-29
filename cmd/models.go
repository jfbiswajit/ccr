package cmd

import (
	"fmt"
	"strings"

	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/openrouter"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models [search]",
	Short: "List available OpenRouter models",
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

	// Header
	fmt.Printf("%-50s %-8s %-14s %-14s\n", "Model ID", "Context", "Input $/M", "Output $/M")
	fmt.Println(strings.Repeat("-", 90))

	for _, m := range models {
		ctx := formatContext(m.ContextLength)
		in := fmt.Sprintf("$%.4f", m.InputCost*1_000_000)
		out := fmt.Sprintf("$%.4f", m.OutputCost*1_000_000)

		id := m.ID
		if query != "" {
			id = highlight(id, query)
		}
		fmt.Printf("%-50s %-8s %-14s %-14s\n", id, ctx, in, out)
	}
	fmt.Printf("\n%d model(s)\n", len(models))
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
	// Bold the matching portion using ANSI
	return s[:idx] + "\033[1m" + s[idx:idx+len(query)] + "\033[0m" + s[idx+len(query):]
}
