package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	URL string
}

// JWTConfig contains JWT-related configuration
type JWTConfig struct {
	SecretKey string
}

// LoadConfig loads the configuration from files and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", "8080")
	v.SetDefault("database.url", "postgres://healthapp_user:verysecretpassword@localhost:5432/healthapp_db?sslmode=disable")
	v.SetDefault("jwt.secretkey", "your-secret-key-change-me-in-production")

	// Set config file paths
	v.AddConfigPath(configPath)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Try to read local config file (for development)
	v.SetConfigName("config.local")
	_ = v.MergeInConfig() // Ignore error if local config doesn't exist

	// Set environment variable prefix and bind environment variables
	v.SetEnvPrefix("HEALTHAPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
