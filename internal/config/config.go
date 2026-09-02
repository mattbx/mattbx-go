// Package config loads runtime configuration from the environment.
//
// In development a .env file is loaded if present; in production Disco injects
// the same keys as real environment variables (see `disco env:set`).
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// minSecretLen is the shortest SESSION_SECRET we accept in production.
// `openssl rand -hex 32` produces 64 characters.
const minSecretLen = 32

type Config struct {
	Env               string // "development" or "production"
	Port              string
	BaseURL           string
	DBPath            string
	AdminPassword     string
	PortfolioPassword string
	SessionSecret     []byte
}

// Development reports whether we're running locally. It gates the Secure flag
// on session cookies, which would otherwise break plain-HTTP localhost.
func (c *Config) Development() bool { return c.Env == "development" }

// Load reads configuration and refuses to start on anything that would be
// silently insecure — a forgeable cookie secret is worse than a crash.
func Load() (*Config, error) {
	// Absent in production; that's expected, so the error is ignored.
	_ = godotenv.Load()

	c := &Config{
		Env:               envOr("ENV", "development"),
		Port:              envOr("PORT", "8080"),
		BaseURL:           strings.TrimRight(envOr("BASE_URL", "http://localhost:8383"), "/"),
		DBPath:            envOr("DB_PATH", "./data/app.db"),
		AdminPassword:     os.Getenv("ADMIN_PASSWORD"),
		PortfolioPassword: os.Getenv("PORTFOLIO_PASSWORD"),
		SessionSecret:     []byte(os.Getenv("SESSION_SECRET")),
	}

	var missing []string
	if c.AdminPassword == "" {
		missing = append(missing, "ADMIN_PASSWORD")
	}
	if c.PortfolioPassword == "" {
		missing = append(missing, "PORTFOLIO_PASSWORD")
	}
	if len(c.SessionSecret) < minSecretLen {
		missing = append(missing, fmt.Sprintf("SESSION_SECRET (need >= %d chars, got %d)", minSecretLen, len(c.SessionSecret)))
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("invalid configuration: %s\n\nSet these in .env locally, or with `disco env:set` in production.", strings.Join(missing, ", "))
	}

	// Sharing one password between the two scopes would hand portfolio
	// visitors write access to the blog.
	if c.AdminPassword == c.PortfolioPassword {
		return nil, fmt.Errorf("ADMIN_PASSWORD and PORTFOLIO_PASSWORD must differ: anyone given the portfolio password would also be able to sign in to /admin")
	}

	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
