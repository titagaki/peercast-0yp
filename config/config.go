// Package config loads application configuration from a TOML file.
package config

import "github.com/BurntSushi/toml"

// Config holds all application configuration.
type Config struct {
	PCP      PCPConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
}

// PCPConfig is PeerCast root server settings.
type PCPConfig struct {
	Addr string // default ":7144"
}

// HTTPConfig is HTTP server settings.
type HTTPConfig struct {
	Addr        string   // default ":80"
	CORSOrigins []string `toml:"cors_origins"`
}

// DatabaseConfig holds MySQL connection parameters.
type DatabaseConfig struct {
	// Full go-sql-driver DSN.
	// Example: "user:pass@tcp(localhost:3306)/peercast?parseTime=true&loc=Local"
	DSN string
}

// Load decodes the TOML file at path and applies defaults for missing fields.
func Load(path string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, err
	}
	applyDefaults(&c)
	return &c, nil
}

func applyDefaults(c *Config) {
	if c.PCP.Addr == "" {
		c.PCP.Addr = ":7144"
	}
	if c.HTTP.Addr == "" {
		c.HTTP.Addr = ":80"
	}
}
