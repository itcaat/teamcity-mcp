package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds the complete server configuration
type Config struct {
	TeamCity TeamCityConfig
	Server   ServerConfig
	Logging  LoggingConfig
	Cache    CacheConfig
}

// TeamCityConfig holds TeamCity connection settings
type TeamCityConfig struct {
	URL     string
	Token   string
	Timeout string
}

// ServerConfig holds server settings
type ServerConfig struct {
	ListenAddr   string
	TLSCert      string
	TLSKey       string
	ServerSecret string
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string
	Format string
}

// CacheConfig holds cache settings
type CacheConfig struct {
	TTL string
}

// Load loads configuration from environment variables only
func Load() (*Config, error) {
	cfg := &Config{
		// Default values
		TeamCity: TeamCityConfig{
			Timeout: getEnvOrDefault("TC_TIMEOUT", "30s"),
		},
		Server: ServerConfig{
			ListenAddr: getEnvOrDefault("LISTEN_ADDR", ":8123"),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		Cache: CacheConfig{
			TTL: getEnvOrDefault("CACHE_TTL", "10s"),
		},
	}

	// Load from environment variables
	loadFromEnv(cfg)

	// Validate required fields
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadFromEnv(cfg *Config) {
	// TeamCity configuration
	cfg.TeamCity.URL = os.Getenv("TC_URL")
	cfg.TeamCity.Token = os.Getenv("TC_TOKEN")

	// Server configuration
	cfg.Server.TLSCert = os.Getenv("TLS_CERT")
	cfg.Server.TLSKey = os.Getenv("TLS_KEY")
	cfg.Server.ServerSecret = os.Getenv("SERVER_SECRET")
}

func validate(cfg *Config) error {
	if cfg.TeamCity.URL == "" {
		return fmt.Errorf("TC_URL environment variable is required")
	}

	if cfg.TeamCity.Token == "" {
		return fmt.Errorf("TC_TOKEN environment variable is required")
	}

	// SERVER_SECRET is now optional - if not provided, authentication will be disabled

	// Validate timeout format
	if _, err := time.ParseDuration(cfg.TeamCity.Timeout); err != nil {
		return fmt.Errorf("invalid TC_TIMEOUT format: %w", err)
	}

	// Validate cache TTL format
	if _, err := time.ParseDuration(cfg.Cache.TTL); err != nil {
		return fmt.Errorf("invalid CACHE_TTL format: %w", err)
	}

	return nil
}

// PrintEnvHelp prints help text for environment variables
func PrintEnvHelp() {
	fmt.Println("TeamCity MCP Server - Environment Variables:")
	fmt.Println()
	fmt.Println("Required:")
	fmt.Println("  TC_URL          TeamCity server URL (e.g., https://your-teamcity-server.com)")
	fmt.Println()
	fmt.Println("Authentication:")
	fmt.Println("  TC_TOKEN        TeamCity API token")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  SERVER_SECRET   Server secret for HMAC token validation (if not set, auth is disabled)")
	fmt.Println("  LISTEN_ADDR     Address to listen on (default: :8123)")
	fmt.Println("  TC_TIMEOUT      HTTP timeout for TeamCity API calls (default: 30s)")
	fmt.Println("  TLS_CERT        Path to TLS certificate file")
	fmt.Println("  TLS_KEY         Path to TLS private key file")
	fmt.Println("  LOG_LEVEL       Log level: debug, info, warn, error (default: info)")
	fmt.Println("  LOG_FORMAT      Log format: json, console (default: json)")
	fmt.Println("  CACHE_TTL       Cache TTL for TeamCity API responses (default: 10s)")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  export TC_URL=https://your-teamcity-server.com")
	fmt.Println("  export TC_TOKEN=your-teamcity-api-token")
	fmt.Println("  # export SERVER_SECRET=your-hmac-secret-key  # Optional - enables auth")
	fmt.Println("  ./server")
}
