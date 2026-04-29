package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/biswajit/ccr/internal/claudesettings"
	"github.com/biswajit/ccr/internal/config"
	"github.com/biswajit/ccr/internal/openrouter"
	"github.com/biswajit/ccr/internal/shell"
	"github.com/biswajit/ccr/internal/statusline"
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

func claudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
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

	// Prompt for API key
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

	// Validate key
	fmt.Print("Validating API key... ")
	client := openrouter.NewClient()
	if err := client.ValidateKey(apiKey); err != nil {
		fmt.Println("failed")
		return fmt.Errorf("invalid API key: %w", err)
	}
	fmt.Println("ok")

	// Snapshot existing statusLine from ~/.claude/settings.json
	settingsPath := claudeSettingsPath()
	existing, err := claudesettings.ReadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("could not read ~/.claude/settings.json: %w", err)
	}
	var originalStatusLine map[string]interface{}
	if sl, ok := existing["statusLine"]; ok {
		if slMap, ok := sl.(map[string]interface{}); ok {
			originalStatusLine = slMap
		}
	}

	// Save config
	if err := config.SaveConfig(dir, &config.Config{
		APIKey:             apiKey,
		ActiveProvider:     "anthropic",
		OriginalStatusLine: originalStatusLine,
	}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Write neutral env file
	envPath := envFilePath()
	if err := shell.WriteEnvFile(envPath, false, ""); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	// Write statusline files
	writeStatuslineFn := statusline.WriteFiles
	if initForce {
		writeStatuslineFn = statusline.WriteFilesForce
	}
	if err := writeStatuslineFn(dir); err != nil {
		return fmt.Errorf("failed to write statusline: %w", err)
	}

	// Print instructions
	shellName, profilePath := shell.DetectShell()
	sourceCmd := shell.SourceInstruction(envPath)

	fmt.Printf("\nConfig saved to ~/.ccr/config.json\n")
	fmt.Printf("Env file written to %s\n", envPath)
	if originalStatusLine != nil {
		fmt.Println("Original statusLine snapshot saved — will be restored on 'ccr disable'")
	}
	fmt.Printf("\nOne-time setup: add this line to your %s (%s):\n\n  %s\n\n", shellName, profilePath, sourceCmd)
	fmt.Println("Statusline scripts written to ~/.ccr/")
	fmt.Println("\nDone. Run 'ccr enable' to switch to OpenRouter.")
	return nil
}

