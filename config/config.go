package config

import (
	"errors"
	"os"
)

// Config holds the configuration for Tinkoff client
type Config struct {
	Token     string
	IsDemo    bool
	ServerURL string
}

// Default server URLs
const (
	ProductionServer = "invest-public-api.tinkoff.ru:443"
	DemoServer       = "sandbox-invest-public-api.tinkoff.ru:443"
)

// New creates a new configuration
func New(token string, isDemo bool) (*Config, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	serverURL := ProductionServer
	if isDemo {
		serverURL = DemoServer
	}

	return &Config{
		Token:     token,
		IsDemo:    isDemo,
		ServerURL: serverURL,
	}, nil
}

// NewFromEnv creates configuration from environment variables
func NewFromEnv() (*Config, error) {
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		return nil, errors.New("TINKOFF_TOKEN environment variable is required")
	}

	isDemo := os.Getenv("TINKOFF_DEMO") == "true"

	return New(token, isDemo)
}
