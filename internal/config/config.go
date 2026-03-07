// Package config loads application configuration from a TOML file.
// Sensitive values (database credentials) are read from environment variables.
package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds all application configuration.
type Config struct {
	PCP      PCPConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
}

// PCPConfig is PeerCast root server settings.
type PCPConfig struct {
	Port             int    `toml:"port"`               // default 7144
	MaxConnections   int    `toml:"max_connections"`    // default 100
	UpdateInterval   int    `toml:"update_interval"`    // seconds; default 120
	HitTimeout       int    `toml:"hit_timeout"`        // seconds; default 180
	MinClientVersion uint32 `toml:"min_client_version"` // default 1200
}

// HTTPConfig is HTTP server settings.
type HTTPConfig struct {
	Port        int      `toml:"port"`         // default 80
	CORSOrigins []string `toml:"cors_origins"`
	YPName      string   `toml:"yp_name"` // displayed in index.txt status line; omit to disable
	YPURL       string   `toml:"yp_url"`  // YP website URL for status line
}

// DatabaseConfig holds database connection parameters.
// These are populated from environment variables, not the TOML file.
type DatabaseConfig struct {
	// DSN is read from the DATABASE_DSN environment variable.
	// Format: "user:pass@tcp(host:port)/dbname?parseTime=true&loc=Local"
	DSN string
}

// Load decodes the TOML file at path, applies defaults, then overlays
// sensitive values from environment variables.
func Load(path string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, err
	}
	applyDefaults(&c)
	applyEnv(&c)
	return &c, nil
}

func applyDefaults(c *Config) {
	if c.PCP.Port == 0 {
		c.PCP.Port = 7144
	}
	if c.PCP.MaxConnections == 0 {
		c.PCP.MaxConnections = 100
	}
	if c.PCP.UpdateInterval == 0 {
		c.PCP.UpdateInterval = 120
	}
	if c.PCP.HitTimeout == 0 {
		c.PCP.HitTimeout = 180
	}
	if c.PCP.MinClientVersion == 0 {
		c.PCP.MinClientVersion = 1200
	}
	if c.HTTP.Port == 0 {
		c.HTTP.Port = 80
	}
}

// applyEnv overlays environment variables onto the config.
// Environment variables take precedence over TOML values.
func applyEnv(c *Config) {
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		c.Database.DSN = v
	}
}
