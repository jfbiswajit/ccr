package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var modelEnvVars = map[string]string{
	"sonnet":  "ANTHROPIC_DEFAULT_SONNET_MODEL",
	"opus":    "ANTHROPIC_DEFAULT_OPUS_MODEL",
	"haiku":   "ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"subagent": "CLAUDE_CODE_SUBAGENT_MODEL",
}

// WriteEnvFile writes ~/.ccr/env.sh (or env.ps1 on Windows).
// enabled=true writes OpenRouter vars; enabled=false writes neutral unsets.
// modelOverrides maps slot name (sonnet/opus/haiku/subagent) to model ID.
func WriteEnvFile(path string, enabled bool, apiKey string, modelOverrides map[string]string) error {
	var content string

	if runtime.GOOS == "windows" {
		content = windowsEnvContent(enabled, apiKey, modelOverrides)
	} else {
		content = unixEnvContent(enabled, apiKey, modelOverrides)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func unixEnvContent(enabled bool, apiKey string, overrides map[string]string) string {
	var b strings.Builder
	b.WriteString("# Managed by ccr — do not edit manually\n")

	if enabled {
		fmt.Fprintf(&b, "export ANTHROPIC_BASE_URL=\"https://openrouter.ai/api\"\n")
		fmt.Fprintf(&b, "export ANTHROPIC_AUTH_TOKEN=\"%s\"\n", apiKey)
		fmt.Fprintf(&b, "export ANTHROPIC_API_KEY=\"\"\n")
	} else {
		b.WriteString("unset ANTHROPIC_BASE_URL\n")
		b.WriteString("unset ANTHROPIC_AUTH_TOKEN\n")
	}

	for slot, modelID := range overrides {
		if envVar, ok := modelEnvVars[slot]; ok {
			fmt.Fprintf(&b, "export %s=\"%s\"\n", envVar, modelID)
		}
	}

	return b.String()
}

func windowsEnvContent(enabled bool, apiKey string, overrides map[string]string) string {
	var b strings.Builder
	b.WriteString("# Managed by ccr — do not edit manually\n")

	if enabled {
		fmt.Fprintf(&b, "$env:ANTHROPIC_BASE_URL = \"https://openrouter.ai/api\"\n")
		fmt.Fprintf(&b, "$env:ANTHROPIC_AUTH_TOKEN = \"%s\"\n", apiKey)
		fmt.Fprintf(&b, "$env:ANTHROPIC_API_KEY = \"\"\n")
	} else {
		b.WriteString("Remove-Item Env:ANTHROPIC_BASE_URL -ErrorAction SilentlyContinue\n")
		b.WriteString("Remove-Item Env:ANTHROPIC_AUTH_TOKEN -ErrorAction SilentlyContinue\n")
	}

	for slot, modelID := range overrides {
		if envVar, ok := modelEnvVars[slot]; ok {
			fmt.Fprintf(&b, "$env:%s = \"%s\"\n", envVar, modelID)
		}
	}

	return b.String()
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
