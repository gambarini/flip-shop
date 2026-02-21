package mcp

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Env var names
const (
	EnvBaseURL   = "FLIP_SHOP_BASE_URL"
	EnvTimeoutMS = "FLIP_SHOP_TIMEOUT_MS"
)

// Defaults from docs/mcp_plan.md
const (
	defaultBaseURL   = "http://localhost:8001"
	defaultTimeoutMS = 8000
)

// Config holds MCP server configuration relevant to proxying flip-shop.
// Additional fields can be added in later phases (logging, bind addr, etc.).
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// LoadFromEnv loads configuration from environment variables, applying sane
// defaults and validation as defined in docs/mcp_plan.md.
func LoadFromEnv() (Config, error) {
	cfg := Config{}

	baseURL := strings.TrimSpace(os.Getenv(EnvBaseURL))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	// Normalize: trim trailing slash to simplify path joins when proxying
	baseURL = strings.TrimRight(baseURL, "/")

	if err := validateBaseURL(baseURL); err != nil {
		return Config{}, err
	}
	cfg.BaseURL = baseURL

	timeoutMSStr := strings.TrimSpace(os.Getenv(EnvTimeoutMS))
	ms := defaultTimeoutMS
	if timeoutMSStr != "" {
		v, err := strconv.Atoi(timeoutMSStr)
		if err != nil {
			return Config{}, errors.New("FLIP_SHOP_TIMEOUT_MS must be an integer (milliseconds)")
		}
		ms = v
	}
	if ms <= 0 {
		return Config{}, errors.New("FLIP_SHOP_TIMEOUT_MS must be > 0")
	}
	cfg.Timeout = time.Duration(ms) * time.Millisecond

	return cfg, nil
}

func validateBaseURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return errors.New("FLIP_SHOP_BASE_URL is not a valid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("FLIP_SHOP_BASE_URL must start with http:// or https://")
	}
	if parsed.Host == "" {
		return errors.New("FLIP_SHOP_BASE_URL must include a host (e.g., localhost:8001)")
	}
	return nil
}
