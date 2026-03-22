package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gitpeekrc")

	content := `{"terminal": "cursor", "ext": ".go,.ts", "exclude": "*_test.go"}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Override HOME to use temp dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()

	if cfg.Terminal != "cursor" {
		t.Errorf("Terminal = %q, want %q", cfg.Terminal, "cursor")
	}
	if cfg.Ext != ".go,.ts" {
		t.Errorf("Ext = %q, want %q", cfg.Ext, ".go,.ts")
	}
	if cfg.Exclude != "*_test.go" {
		t.Errorf("Exclude = %q, want %q", cfg.Exclude, "*_test.go")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()

	if cfg.Terminal != "" || cfg.Ext != "" || cfg.Exclude != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gitpeekrc")

	if err := os.WriteFile(configPath, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()

	// Should return empty config on parse error
	if cfg.Terminal != "" || cfg.Ext != "" || cfg.Exclude != "" {
		t.Errorf("expected empty config on invalid JSON, got %+v", cfg)
	}
}

func TestLoadConfig_PartialFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gitpeekrc")

	content := `{"terminal": "zed"}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()

	if cfg.Terminal != "zed" {
		t.Errorf("Terminal = %q, want %q", cfg.Terminal, "zed")
	}
	if cfg.Ext != "" {
		t.Errorf("Ext = %q, want empty", cfg.Ext)
	}
	if cfg.Exclude != "" {
		t.Errorf("Exclude = %q, want empty", cfg.Exclude)
	}
}
