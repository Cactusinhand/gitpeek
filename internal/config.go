package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds persistent user preferences loaded from ~/.gitpeekrc
type Config struct {
	Terminal string `json:"terminal,omitempty"`
	Ext      string `json:"ext,omitempty"`
	Exclude  string `json:"exclude,omitempty"`
}

// LoadConfig reads config from ~/.gitpeekrc. Returns empty config if file doesn't exist.
func LoadConfig() Config {
	var cfg Config
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(filepath.Join(home, ".gitpeekrc"))
	if err != nil {
		return cfg
	}

	json.Unmarshal(data, &cfg)
	return cfg
}
