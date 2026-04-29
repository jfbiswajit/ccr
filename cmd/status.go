package cmd

import (
	"fmt"
	"os"

	"github.com/biswajit/ccr/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current provider and configuration state",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir := ccrDir()
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		fmt.Println("ccr is not initialized. Run 'ccr init' first.")
		return nil
	}

	provider := cfg.ActiveProvider
	if provider == "openrouter" {
		fmt.Println("Provider: OpenRouter (active)")
	} else {
		fmt.Println("Provider: Anthropic (active)")
	}

	envPath := envFilePath()
	fmt.Printf("Env file:  %s\n", envPath)

	// Detect if env vars are actually sourced in current shell
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if provider == "openrouter" && baseURL == "https://openrouter.ai/api" {
		fmt.Println("Shell:     env vars sourced ✓")
	} else if provider == "anthropic" && baseURL == "" {
		fmt.Println("Shell:     env vars sourced ✓")
	} else {
		fmt.Printf("Shell:     not yet sourced — run: source %s\n", envPath)
	}

	return nil
}
