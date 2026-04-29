package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// WriteEnvFile writes ~/.ccr/env.sh (or env.ps1 on Windows).
// enabled=true writes OpenRouter vars; enabled=false writes neutral unsets.
func WriteEnvFile(path string, enabled bool, apiKey string) error {
	var content string

	if runtime.GOOS == "windows" {
		content = windowsEnvContent(enabled, apiKey)
	} else {
		content = unixEnvContent(enabled, apiKey)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func unixEnvContent(enabled bool, apiKey string) string {
	if enabled {
		return fmt.Sprintf(`# Managed by ccr — do not edit manually
export ANTHROPIC_BASE_URL="https://openrouter.ai/api"
export ANTHROPIC_AUTH_TOKEN="%s"
export ANTHROPIC_API_KEY=""
`, apiKey)
	}
	return `# Managed by ccr — do not edit manually
unset ANTHROPIC_BASE_URL
unset ANTHROPIC_AUTH_TOKEN
`
}

func windowsEnvContent(enabled bool, apiKey string) string {
	if enabled {
		return fmt.Sprintf(`# Managed by ccr — do not edit manually
$env:ANTHROPIC_BASE_URL = "https://openrouter.ai/api"
$env:ANTHROPIC_AUTH_TOKEN = "%s"
$env:ANTHROPIC_API_KEY = ""
`, apiKey)
	}
	return `# Managed by ccr — do not edit manually
Remove-Item Env:ANTHROPIC_BASE_URL -ErrorAction SilentlyContinue
Remove-Item Env:ANTHROPIC_AUTH_TOKEN -ErrorAction SilentlyContinue
`
}

// DetectShell returns the shell name and the recommended profile file path.
func DetectShell() (name string, profile string) {
	home, _ := os.UserHomeDir()

	if runtime.GOOS == "windows" {
		return "powershell", filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	}

	shellEnv := os.Getenv("SHELL")
	base := filepath.Base(shellEnv)

	switch base {
	case "zsh":
		return "zsh", filepath.Join(home, ".zshrc")
	case "bash":
		// macOS bash uses .bash_profile; Linux uses .bashrc
		if runtime.GOOS == "darwin" {
			return "bash", filepath.Join(home, ".bash_profile")
		}
		return "bash", filepath.Join(home, ".bashrc")
	case "fish":
		return "fish", filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return strings.TrimPrefix(base, ""), filepath.Join(home, ".profile")
	}
}

// SourceInstruction returns the line the user needs to add to their shell profile.
func SourceInstruction(envFilePath string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`. "%s"`, envFilePath)
	}
	return fmt.Sprintf(`source "%s"`, envFilePath)
}
