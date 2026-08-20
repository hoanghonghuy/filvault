package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the persisted CLI state (tokens + API base).
type Config struct {
	APIBase      string `json:"apiBase"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// DefaultAPIBase is used when FILVAULT_API_BASE is unset.
const DefaultAPIBase = "http://localhost:8080/api/v1"

// Path returns the config file path for the current OS.
// FILVAULT_CONFIG_PATH overrides the default (used by tests).
func Path() string {
	if override := os.Getenv("FILVAULT_CONFIG_PATH"); override != "" {
		return override
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "filvault", "config.json")
	}
	return filepath.Join(".", "filvault-config.json")
}

// Load reads the config file. A missing file returns an empty Config.
func Load() (Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes the config file, creating parent directories as needed.
func Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
