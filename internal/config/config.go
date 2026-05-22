// Package config handles application configuration and environment variables.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads variables from a local .env file if present.
// It is safe to call multiple times; subsequent calls are no-ops.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		// .env file is optional in many environments (e.g., CI/CD)
		_ = 0
	}
}

// APIKey returns the Etherscan API key from the environment.
func APIKey() string {
	return os.Getenv("ETHERSCAN_API_KEY")
}
