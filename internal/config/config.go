package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for vaultshift.
type Config struct {
	SourceAddr      string        `mapstructure:"source_addr"`
	SourceToken     string        `mapstructure:"source_token"`
	SourceNamespace string        `mapstructure:"source_namespace"`
	DestAddr        string        `mapstructure:"dest_addr"`
	DestToken       string        `mapstructure:"dest_token"`
	DestNamespace   string        `mapstructure:"dest_namespace"`
	DryRun          bool          `mapstructure:"dry_run"`
	AuditLogPath    string        `mapstructure:"audit_log_path"`
	Timeout         time.Duration `mapstructure:"timeout"`
}

// Load reads configuration from file and environment variables.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("timeout", 30*time.Second)
	v.SetDefault("dry_run", false)
	v.SetDefault("audit_log_path", "vaultshift-audit.log")

	v.SetEnvPrefix("VAULTSHIFT")
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName(".vaultshift")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath(os.ExpandEnv("$HOME"))
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.SourceAddr == "" {
		return fmt.Errorf("source_addr is required")
	}
	if c.SourceToken == "" {
		return fmt.Errorf("source_token is required")
	}
	if c.DestAddr == "" {
		return fmt.Errorf("dest_addr is required")
	}
	if c.DestToken == "" {
		return fmt.Errorf("dest_token is required")
	}
	return nil
}
