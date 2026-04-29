package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteStatuslineFiles(t *testing.T) {
	dir := t.TempDir()

	if err := WriteFiles(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tsPath := filepath.Join(dir, "statusline.ts")
	shPath := filepath.Join(dir, "statusline.sh")

	// Both files must exist
	if _, err := os.Stat(tsPath); err != nil {
		t.Errorf("statusline.ts missing: %v", err)
	}
	if _, err := os.Stat(shPath); err != nil {
		t.Errorf("statusline.sh missing: %v", err)
	}

	// .ts must contain OpenRouter fetch logic
	tsData, _ := os.ReadFile(tsPath)
	if !strings.Contains(string(tsData), "openrouter.ai") {
		t.Error("statusline.ts missing openrouter.ai reference")
	}
	if !strings.Contains(string(tsData), "total_cost") {
		t.Error("statusline.ts missing cost tracking logic")
	}

	// .sh must be executable and reference the .ts file
	shData, _ := os.ReadFile(shPath)
	if !strings.Contains(string(shData), "statusline.ts") {
		t.Error("statusline.sh should reference statusline.ts")
	}
	shInfo, _ := os.Stat(shPath)
	if shInfo.Mode()&0111 == 0 {
		t.Error("statusline.sh should be executable")
	}
}

func TestWriteFilesSkipsIfExists(t *testing.T) {
	dir := t.TempDir()

	// Write once
	if err := WriteFiles(dir); err != nil {
		t.Fatalf("first write error: %v", err)
	}

	// Modify the file
	tsPath := filepath.Join(dir, "statusline.ts")
	os.WriteFile(tsPath, []byte("// modified"), 0644)

	// Write again without force — should skip
	if err := WriteFiles(dir); err != nil {
		t.Fatalf("second write error: %v", err)
	}

	data, _ := os.ReadFile(tsPath)
	if string(data) != "// modified" {
		t.Error("WriteFiles should not overwrite existing files without force")
	}
}

func TestWriteFilesForce(t *testing.T) {
	dir := t.TempDir()

	if err := WriteFiles(dir); err != nil {
		t.Fatalf("first write error: %v", err)
	}

	tsPath := filepath.Join(dir, "statusline.ts")
	os.WriteFile(tsPath, []byte("// modified"), 0644)

	if err := WriteFilesForce(dir); err != nil {
		t.Fatalf("force write error: %v", err)
	}

	data, _ := os.ReadFile(tsPath)
	if string(data) == "// modified" {
		t.Error("WriteFilesForce should overwrite existing files")
	}
}
