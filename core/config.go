package core

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type RadConfig struct {
	InvocationLogging *InvocationLoggingConfig `toml:"invocation_logging"`
}

type InvocationLoggingConfig struct {
	Enabled        bool `toml:"enabled"`
	IncludeArgs    bool `toml:"include_args"`
	MaxSizeMB      int  `toml:"max_size_mb"` // make float64?
	KeepRolledLogs int  `toml:"keep_rolled_logs"`
}

func DefaultRadConfig() *RadConfig {
	return &RadConfig{
		InvocationLogging: defaultInvocationLoggingConfig(),
	}
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

	// Validate and fix individual fields
	invocationLoggingDefaults := defaultInvocationLoggingConfig()

	if config.InvocationLogging.MaxSizeMB <= 0 {
		warnf(configPath, "Invalid config: max_size_mb must be > 0, got %d. Using default: %d MB\n",
			config.InvocationLogging.MaxSizeMB, invocationLoggingDefaults.MaxSizeMB)
		config.InvocationLogging.MaxSizeMB = invocationLoggingDefaults.MaxSizeMB
	}

	if config.InvocationLogging.KeepRolledLogs < 0 {
		warnf(configPath, "Invalid config: keep_rolled_logs must be >= 0, got %d. Using default: %d\n",
			config.InvocationLogging.KeepRolledLogs, invocationLoggingDefaults.KeepRolledLogs)
		config.InvocationLogging.KeepRolledLogs = invocationLoggingDefaults.KeepRolledLogs
	}

	return config
}

func defaultInvocationLoggingConfig() *InvocationLoggingConfig {
	return &InvocationLoggingConfig{
		Enabled:        false, // opt-in
		IncludeArgs:    false, // args may contain sensitive info, make opt-in
		MaxSizeMB:      10,
		KeepRolledLogs: 2,
	}
}

// TODO as of writing, this might get invoked before global flags e.g. 'Quiet' are set
// what we should *probably* do is accumulate warnings and print them later
func warnf(configPath string, format string, args ...interface{}) {
	RP.RadStderrf("Warning! "+format, args...)
	RP.RadStderrf("Please fix the invalid config file: %s\n", configPath)
}
