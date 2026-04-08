package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all CLI configuration.
type Config struct {
	ClientID string `yaml:"client_id"`
	APIKey   string `yaml:"api_key"`
	Env      string `yaml:"env"`
	Output   string `yaml:"output"`
	Debug    bool   `yaml:"-"` // runtime-only, never written to config file
}

// DefaultPath returns the path to the config file (~/.uqpay/config.yaml).
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".uqpay", "config.yaml")
}

// TokenCachePath returns the path to the token cache (~/.uqpay/token.json).
func TokenCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".uqpay", "token.json")
}

// Load reads the config file and returns a Config with defaults applied.
// If the file does not exist, defaults are returned without error.
func Load() (*Config, error) {
	cfg := &Config{Env: "sandbox", Output: "table"}
	path := DefaultPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Env == "" {
		cfg.Env = "sandbox"
	}
	if cfg.Output == "" {
		cfg.Output = "table"
	}
	return cfg, nil
}

// ApplyEnvVars overrides config fields with UQPAY_* environment variables.
func (c *Config) ApplyEnvVars() {
	if v := os.Getenv("UQPAY_CLIENT_ID"); v != "" {
		c.ClientID = v
	}
	if v := os.Getenv("UQPAY_API_KEY"); v != "" {
		c.APIKey = v
	}
	if v := os.Getenv("UQPAY_ENV"); v != "" {
		c.Env = v
	}
	if v := os.Getenv("UQPAY_OUTPUT"); v != "" {
		c.Output = v
	}
}

// Save writes the config to ~/.uqpay/config.yaml (mode 0600).
func (c *Config) Save() error {
	path := DefaultPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Set updates a single config key by name and saves.
func (c *Config) Set(key, value string) error {
	switch key {
	case "client-id":
		c.ClientID = value
	case "api-key":
		c.APIKey = value
	case "env":
		if value != "sandbox" && value != "production" {
			return &invalidValue{key: key, value: value, allowed: "sandbox, production"}
		}
		c.Env = value
	case "output":
		if value != "table" && value != "json" && value != "yaml" {
			return &invalidValue{key: key, value: value, allowed: "table, json, yaml"}
		}
		c.Output = value
	default:
		return &unknownKey{key: key}
	}
	return c.Save()
}

type unknownKey struct{ key string }

func (e *unknownKey) Error() string {
	return "unknown config key: " + e.key + " (valid: client-id, api-key, env, output)"
}

type invalidValue struct{ key, value, allowed string }

func (e *invalidValue) Error() string {
	return "invalid value " + e.value + " for " + e.key + " (allowed: " + e.allowed + ")"
}
