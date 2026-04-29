package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/openrouter"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up ccr with your OpenRouter API key",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config")
	rootCmd.AddCommand(initCmd)
}

func ccrDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ccr")
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := ccrDir()
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return err
	}

	if cfg.APIKey != "" && !initForce {
		fmt.Println("ccr is already initialized. Run with --force to overwrite.")
		return nil
	}

	var apiKey string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("OpenRouter API key").
				Description("Get yours at openrouter.ai/keys").
				Password(true).
				Value(&apiKey).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API key cannot be empty")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	fmt.Print("Validating API key... ")
	client := openrouter.NewClient()
	if err := client.ValidateKey(apiKey); err != nil {
		fmt.Println("failed")
		return fmt.Errorf("invalid API key: %w", err)
	}
	fmt.Println("ok")

	if err := config.SaveConfig(dir, &config.Config{
		APIKey:         apiKey,
		ActiveProvider: "anthropic",
	}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\nConfig saved to ~/.ccr/config.json")
	fmt.Println("\nNext steps will be completed in subsequent init stages.")
	return nil
}
