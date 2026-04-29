package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()
	out := buf.String()
	if !strings.Contains(out, "ccr") {
		t.Errorf("help output missing 'ccr', got: %s", out)
	}
	if !strings.Contains(out, "init") {
		t.Errorf("help output missing 'init' subcommand, got: %s", out)
	}
}

func TestRootVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--version"})
	_ = rootCmd.Execute()
	out := buf.String()
	if !strings.Contains(out, "ccr") {
		t.Errorf("version output missing 'ccr', got: %s", out)
	}
}
