package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILVAULT_CONFIG_PATH", filepath.Join(dir, "config.json"))

	cfg := Config{
		APIBase:      "http://example.com/api/v1",
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.APIBase != cfg.APIBase || got.AccessToken != cfg.AccessToken || got.RefreshToken != cfg.RefreshToken {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", got, cfg)
	}
}

func TestLoadMissingReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILVAULT_CONFIG_PATH", filepath.Join(dir, "nope.json"))

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.APIBase != "" || got.AccessToken != "" || got.RefreshToken != "" {
		t.Fatalf("expected empty config, got %+v", got)
	}
}

func TestPathUsesEnvOverride(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "custom.json")
	t.Setenv("FILVAULT_CONFIG_PATH", override)

	if got := Path(); got != override {
		t.Fatalf("Path() = %q, want %q", got, override)
	}
}

func TestSaveCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "config.json")
	t.Setenv("FILVAULT_CONFIG_PATH", nested)

	if err := Save(Config{APIBase: "x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}
