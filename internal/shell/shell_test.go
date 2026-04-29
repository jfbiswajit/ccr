package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteNeutralEnvSh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	if err := WriteEnvFile(path, false, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "unset ANTHROPIC_BASE_URL") {
		t.Errorf("neutral env.sh missing 'unset ANTHROPIC_BASE_URL'")
	}
	if !strings.Contains(content, "unset ANTHROPIC_AUTH_TOKEN") {
		t.Errorf("neutral env.sh missing 'unset ANTHROPIC_AUTH_TOKEN'")
	}
	if strings.Contains(content, "export ANTHROPIC_BASE_URL") {
		t.Errorf("neutral env.sh should not export ANTHROPIC_BASE_URL")
	}
}

func TestWriteEnabledEnvSh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	if err := WriteEnvFile(path, true, "sk-or-test-key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, `export ANTHROPIC_BASE_URL="https://openrouter.ai/api"`) {
		t.Errorf("enabled env.sh missing ANTHROPIC_BASE_URL export")
	}
	if !strings.Contains(content, `export ANTHROPIC_AUTH_TOKEN="sk-or-test-key"`) {
		t.Errorf("enabled env.sh missing ANTHROPIC_AUTH_TOKEN export")
	}
	if !strings.Contains(content, `export ANTHROPIC_API_KEY=""`) {
		t.Errorf("enabled env.sh missing blank ANTHROPIC_API_KEY")
	}
}

func TestDetectShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	shell, profile := DetectShell()
	if shell != "zsh" {
		t.Errorf("expected zsh, got %s", shell)
	}
	if !strings.Contains(profile, ".zshrc") {
		t.Errorf("expected .zshrc profile, got %s", profile)
	}
}

func TestDetectShellBash(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	shell, profile := DetectShell()
	if shell != "bash" {
		t.Errorf("expected bash, got %s", shell)
	}
	if !strings.Contains(profile, ".bashrc") && !strings.Contains(profile, ".bash_profile") {
		t.Errorf("expected bash profile, got %s", profile)
	}
}

func TestDetectShellFallback(t *testing.T) {
	t.Setenv("SHELL", "")
	shell, profile := DetectShell()
	if shell == "" {
		t.Error("expected fallback shell name")
	}
	if profile == "" {
		t.Error("expected fallback profile")
	}
}
