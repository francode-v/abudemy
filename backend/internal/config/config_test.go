package config

import (
	"testing"
	"time"
)

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("APP_ENV", "development")
}

func TestLoadAppliesDefaults(t *testing.T) {
	setBaseEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("addr = %s", cfg.Addr)
	}
	if cfg.JWTAccessTTL != 15*time.Minute {
		t.Fatalf("access ttl = %s", cfg.JWTAccessTTL)
	}
	if cfg.NotifyProvider != "none" {
		t.Fatalf("notify provider = %s", cfg.NotifyProvider)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error when DATABASE_URL is missing")
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error when JWT_SECRET is missing")
	}
}

func TestProductionRequiresLongSecret(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a short secret in production")
	}
}

func TestWebhookRequiresURL(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("NOTIFY_PROVIDER", "webhook")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error when webhook provider has no URL")
	}
}

func TestParsesEnvironmentOverrides(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("BCRYPT_COST", "11")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Addr != ":9090" || cfg.BcryptCost != 11 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("cors origins = %v", cfg.CORSAllowedOrigins)
	}
}

func TestResolveAddrUsesPortWhenNoHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("PORT", "10000")
	if got := resolveAddr(); got != ":10000" {
		t.Fatalf("resolveAddr() = %q, want :10000", got)
	}
	t.Setenv("HTTP_ADDR", ":9999")
	if got := resolveAddr(); got != ":9999" {
		t.Fatalf("HTTP_ADDR must win, got %q", got)
	}
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("PORT", "")
	if got := resolveAddr(); got != ":8080" {
		t.Fatalf("default = %q, want :8080", got)
	}
}
