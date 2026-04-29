package cmd

import (
	"path/filepath"
	"runtime"
)

func envFileNameForOS() string {
	if runtime.GOOS == "windows" {
		return "env.ps1"
	}
	return "env.sh"
}

func envFilePath() string {
	return filepath.Join(ccrDir(), envFileNameForOS())
}
