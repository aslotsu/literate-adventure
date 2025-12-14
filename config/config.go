package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// NATS Configuration
	NatsURL         string
	NatsCredsFile   string

	// DynamoDB Configuration
	AWSRegion       string
	NotifTableName  string

	// AWS Credentials (optional if using IAM roles)
	AWSAccessKeyID  string
	AWSSecretKey    string

	// Pusher Configuration (for real-time notifications)
	PusherAppID     string
	PusherKey       string
	PusherSecret    string
	PusherCluster   string

	// Application Configuration
	Environment     string
	LogLevel        string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Try to load .env file (optional in production)
	_ = godotenv.Load()

	config := &Config{
		NatsURL:        getEnv("NATS_URL", "nats://connect.ngs.global"),
		NatsCredsFile:  getEnv("NATS_CREDS_FILE", "NGS-Default-exobook.creds"),
		AWSRegion:      getEnv("AWS_REGION", "ca-central-1"),
		NotifTableName: getEnv("NOTIF_TABLE_NAME", "exobook-notifications"),
		AWSAccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		PusherAppID:    os.Getenv("PUSHER_APP_ID"),
		PusherKey:      getEnv("PUSHER_KEY", "a77d99a67f8892897039"),
		PusherSecret:   os.Getenv("PUSHER_SECRET"),
		PusherCluster:  getEnv("PUSHER_CLUSTER", "mt1"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if config.NatsURL == "" {
		return nil, fmt.Errorf("NATS_URL is required")
	}

	if config.AWSRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	if config.NotifTableName == "" {
		return nil, fmt.Errorf("NOTIF_TABLE_NAME is required")
	}

	return config, nil
}

// getEnv gets an environment variable with a fallback default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
