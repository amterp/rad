package core

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type RadConfig struct {
}

func DefaultRadConfig() *RadConfig {
	return &RadConfig{}
}

// LoadRadConfig loads the invocation logging configuration from ~/.rad/config.toml (configurable)
// Returns defaults if file doesn't exist or on parse errors (with warnings)
func LoadRadConfig() *RadConfig {
	configPath := filepath.Join(RadHomeInst.HomeDir, "config.toml")

	// Start from default, so when we load, unspecified fields remain defaults (instead of nil)
	config := DefaultRadConfig()

	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			warnf(configPath, "Using default config; cannot read config file: %v\n", err)
		}
		// If file doesn't exist, use defaults silently
		return config
	}

	// Try to load and parse
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		warnf(configPath, "Using default config; failed to parse config file: %v\n", err)
		return config
	}

	return config
}

// TODO as of writing, this might get invoked before global flags e.g. 'Quiet' are set
// what we should *probably* do is accumulate warnings and print them later
func warnf(configPath string, format string, args ...interface{}) {
	RP.RadStderrf("Warning! "+format, args...)
	RP.RadStderrf("Please fix the invalid config file: %s\n", configPath)
}
