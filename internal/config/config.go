package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for a vaultshift run.
type Config struct {
	SourceAddress   string   `yaml:"source_address"   env:"VAULT_SRC_ADDR"`
	SourceToken     string   `yaml:"source_token"     env:"VAULT_SRC_TOKEN"`
	SourceNamespace string   `yaml:"source_namespace" env:"VAULT_SRC_NAMESPACE"`
	DestAddress     string   `yaml:"dest_address"     env:"VAULT_DEST_ADDR"`
	DestToken       string   `yaml:"dest_token"       env:"VAULT_DEST_TOKEN"`
	DestNamespace   string   `yaml:"dest_namespace"   env:"VAULT_DEST_NAMESPACE"`
	Prefix          string   `yaml:"prefix"           env:"VAULT_PREFIX"`
	DryRun          bool     `yaml:"dry_run"`
	AuditLog        string   `yaml:"audit_log"        env:"VAULT_AUDIT_LOG"`
	IncludePrefixes []string `yaml:"include_prefixes"`
	ExcludePrefixes []string `yaml:"exclude_prefixes"`
}

// Load reads a YAML config file and overlays environment variables.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	overrideFromEnv(cfg)
	return cfg, Validate(cfg)
}

func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("VAULT_SRC_ADDR"); v != "" {
		cfg.SourceAddress = v
	}
	if v := os.Getenv("VAULT_SRC_TOKEN"); v != "" {
		cfg.SourceToken = v
	}
	if v := os.Getenv("VAULT_SRC_NAMESPACE"); v != "" {
		cfg.SourceNamespace = v
	}
	if v := os.Getenv("VAULT_DEST_ADDR"); v != "" {
		cfg.DestAddress = v
	}
	if v := os.Getenv("VAULT_DEST_TOKEN"); v != "" {
		cfg.DestToken = v
	}
	if v := os.Getenv("VAULT_DEST_NAMESPACE"); v != "" {
		cfg.DestNamespace = v
	}
	if v := os.Getenv("VAULT_PREFIX"); v != "" {
		cfg.Prefix = v
	}
	if v := os.Getenv("VAULT_AUDIT_LOG"); v != "" {
		cfg.AuditLog = v
	}
}

// Validate returns an error if required fields are missing.
func Validate(cfg *Config) error {
	if cfg.SourceAddress == "" {
		return errors.New("source_address is required")
	}
	if cfg.SourceToken == "" {
		return errors.New("source_token is required")
	}
	if cfg.DestAddress == "" {
		return errors.New("dest_address is required")
	}
	if cfg.DestToken == "" {
		return errors.New("dest_token is required")
	}
	return nil
}
