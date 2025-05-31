package config

import (
	"os"
)

// Config holds all the configuration for the application
type Config struct {
	// Server configuration
	Port string

	// AWS configuration
	AWSRegion    string
	DynamoDBEndpoint string

	// Security
	JWTSecret string
	JWTExpirationHours int
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() Config {
	return Config{
		// Server configuration
		Port: getEnv("PORT", "8080"),

		// AWS configuration
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
		DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", ""),

		// Security
		JWTSecret: getEnv("JWT_SECRET", "dev-reserve-secret-key"),
		JWTExpirationHours: 24,
	}
}

// getEnv retrieves an environment variable or returns a default value if not found
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
