package config

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads variables from a local .env file if present.
// It is safe to call multiple times; subsequent calls are no-ops.
func LoadEnv() {
	_ = godotenv.Load()
}

// APIKey returns the Etherscan API key from the environment.
func APIKey() string {
	return os.Getenv("ETHERSCAN_API_KEY")
}
