// Package config loads and validates application configuration from the
// environment. No secret values are ever logged or returned through the API.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment identifies the runtime environment.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// Config holds every runtime setting for the API.
type Config struct {
	Env  Environment
	Addr string

	DatabaseURL string
	// DBMaxConns caps the pgx pool size (DB_MAX_CONNS, default 10).
	DBMaxConns int
	// SeedDemoData applies the demo migrations (SEED_DEMO_DATA). Defaults to
	// true only in development; demo accounts use publicly known passwords.
	SeedDemoData bool

	JWTSecret      string
	JWTIssuer      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	PasswordReset  time.Duration
	BcryptCost     int
	MaxLoginErrors int

	LogLevel string

	NotifyProvider    string
	NotifyWebhookURL  string
	NotifyWebhookAuth string

	AssistantProvider    string
	AssistantWebhookURL  string
	AssistantWebhookAuth string

	GeoProvider  string
	GeoBaseURL   string
	GeoUserAgent string

	PaymentsProvider string

	CORSAllowedOrigins []string

	// SeedAdminEmail / SeedAdminPassword are only used by an optional bootstrap
	// command. They are never required for the API to run.
	SeedAdminEmail    string
	SeedAdminPassword string
}

// IsProduction reports whether the API runs in a production environment.
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

// IsDevelopment reports whether the API runs in a development environment.
func (c *Config) IsDevelopment() bool { return c.Env == EnvDevelopment }

// Load reads configuration from the process environment, applying safe
// defaults for development. Required secrets must be provided explicitly.
func Load() (*Config, error) {
	env := Environment(getEnv("APP_ENV", string(EnvDevelopment)))
	cfg := &Config{
		Env:                  env,
		Addr:                 resolveAddr(),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		DBMaxConns:           getEnvInt("DB_MAX_CONNS", 10),
		SeedDemoData:         getEnvBool("SEED_DEMO_DATA", env == EnvDevelopment),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTIssuer:            getEnv("JWT_ISSUER", "abudemy"),
		JWTRefreshTTL:        getEnvDuration("JWT_REFRESH_TTL", 720*time.Hour),
		JWTAccessTTL:         getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
		PasswordReset:        getEnvDuration("PASSWORD_RESET_TTL", time.Hour),
		BcryptCost:           getEnvInt("BCRYPT_COST", 12),
		MaxLoginErrors:       getEnvInt("MAX_LOGIN_ATTEMPTS", 0),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		NotifyProvider:       strings.ToLower(getEnv("NOTIFY_PROVIDER", "none")),
		NotifyWebhookURL:     getEnv("NOTIFY_WEBHOOK_URL", ""),
		NotifyWebhookAuth:    getEnv("NOTIFY_WEBHOOK_TOKEN", ""),
		AssistantProvider:    strings.ToLower(getEnv("ASSISTANT_PROVIDER", "none")),
		AssistantWebhookURL:  getEnv("ASSISTANT_WEBHOOK_URL", ""),
		AssistantWebhookAuth: getEnv("ASSISTANT_WEBHOOK_TOKEN", ""),
		GeoProvider:          strings.ToLower(getEnv("GEOCODE_PROVIDER", "none")),
		GeoBaseURL:           getEnv("GEOCODE_BASE_URL", ""),
		GeoUserAgent:         getEnv("GEOCODE_USER_AGENT", ""),
		PaymentsProvider:     strings.ToLower(getEnv("PAYMENTS_PROVIDER", "none")),
		CORSAllowedOrigins:   getEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
		SeedAdminEmail:       getEnv("SEED_ADMIN_EMAIL", ""),
		SeedAdminPassword:    getEnv("SEED_ADMIN_PASSWORD", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	switch c.Env {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("config: invalid APP_ENV %q", c.Env)
	}

	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("config: DATABASE_URL is required")
	}
	if strings.TrimSpace(c.JWTSecret) == "" {
		return errors.New("config: JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 && !c.IsDevelopment() {
		return errors.New("config: JWT_SECRET must be at least 32 bytes outside development")
	}
	if c.DBMaxConns < 1 || c.DBMaxConns > 100 {
		return fmt.Errorf("config: DB_MAX_CONNS must be between 1 and 100, got %d", c.DBMaxConns)
	}
	if c.BcryptCost < 10 || c.BcryptCost > 16 {
		return fmt.Errorf("config: BCRYPT_COST must be between 10 and 16, got %d", c.BcryptCost)
	}
	if c.JWTAccessTTL <= 0 || c.JWTRefreshTTL <= 0 {
		return errors.New("config: token TTLs must be positive")
	}
	switch c.NotifyProvider {
	case "none", "log", "webhook":
	default:
		return fmt.Errorf("config: NOTIFY_PROVIDER must be one of none|log|webhook, got %q", c.NotifyProvider)
	}
	if c.NotifyProvider == "webhook" && strings.TrimSpace(c.NotifyWebhookURL) == "" {
		return errors.New("config: NOTIFY_WEBHOOK_URL is required when NOTIFY_PROVIDER=webhook")
	}
	switch c.AssistantProvider {
	case "none", "webhook":
	default:
		return fmt.Errorf("config: ASSISTANT_PROVIDER must be one of none|webhook, got %q", c.AssistantProvider)
	}
	if c.AssistantProvider == "webhook" && strings.TrimSpace(c.AssistantWebhookURL) == "" {
		return errors.New("config: ASSISTANT_WEBHOOK_URL is required when ASSISTANT_PROVIDER=webhook")
	}
	switch c.GeoProvider {
	case "none", "nominatim":
	default:
		return fmt.Errorf("config: GEOCODE_PROVIDER must be one of none|nominatim, got %q", c.GeoProvider)
	}
	switch c.PaymentsProvider {
	case "none":
	default:
		return fmt.Errorf("config: PAYMENTS_PROVIDER must be none (no payment provider is available), got %q", c.PaymentsProvider)
	}
	return nil
}

// resolveAddr returns the listen address. HTTP_ADDR wins; otherwise the PORT
// variable injected by platforms such as Render is used; otherwise :8080.
func resolveAddr() string {
	if v := strings.TrimSpace(os.Getenv("HTTP_ADDR")); v != "" {
		return v
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}
	return ":8080"
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
