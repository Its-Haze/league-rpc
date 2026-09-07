package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// AppName is the application name used for config directory
	AppName = "league-rpc"
	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.json"
)

// GetConfigDir returns the configuration directory path
// Windows: %APPDATA%\league-rpc
func GetConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}

	configDir := filepath.Join(appData, AppName)
	return configDir, nil
}

// GetConfigPath returns the full path to the configuration file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, ConfigFileName), nil
}

// Load reads the configuration from disk
// If the file doesn't exist, it returns the default configuration
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// An older file is walked up to the current schema before it is parsed.
	data, migrated, err := migrateToCurrent(data)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Repair anything out of bounds so a hand-edited file still boots.
	config.clamp()

	if migrated {
		config.SchemaVersion = CurrentSchemaVersion
		_ = Save(&config) // best-effort upgrade; a read-only dir shouldn't block startup
	}

	return &config, nil
}

// Save writes the configuration to disk
func Save(config *Config) error {
	// Validate before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadOrCreate loads the configuration from disk, or creates a default one if it doesn't exist
func LoadOrCreate() (*Config, error) {
	config, err := Load()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist on disk, save the default config
	configPath, _ := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := Save(config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return config, nil
}
