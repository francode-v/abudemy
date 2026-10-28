// Command api is the Abudemy REST backend entrypoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/abudemy/backend/internal/api"
	"github.com/abudemy/backend/internal/assistant"
	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/config"
	"github.com/abudemy/backend/internal/database"
	"github.com/abudemy/backend/internal/geo"
	"github.com/abudemy/backend/internal/logging"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/notify"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load() // .env is optional; real environments use process env.

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel, cfg.IsProduction())
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	migrateCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	if err := database.Migrate(migrateCtx, pool, database.Options{SeedDemo: cfg.SeedDemoData}); err != nil {
		return err
	}
	logger.Info("database ready", "demo_data", cfg.SeedDemoData)
	if cfg.SeedDemoData && cfg.IsProduction() {
		logger.Warn("SEED_DEMO_DATA is enabled in production: demo accounts with publicly known passwords exist")
	}
	if cfg.IsProduction() {
		for _, o := range cfg.CORSAllowedOrigins {
			if o == "*" {
				logger.Warn("CORS_ALLOWED_ORIGINS is '*' in production; restrict it to your frontend origin")
				break
			}
		}
	}

	st := store.New(pool)
	if err := seedAdmin(ctx, logger, st, cfg); err != nil {
		return err
	}

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessTTL)
	hasher := auth.NewPasswordHasher(cfg.BcryptCost)
	notifier := notify.New(cfg.NotifyProvider, cfg.NotifyWebhookURL, cfg.NotifyWebhookAuth, logger)
	if notifier.Name() == "none" {
		logger.Warn("no notification provider configured; assistance notifications will be recorded as skipped")
	}

	services := service.New(service.Deps{
		Store:            st,
		Tokens:           tokens,
		Hasher:           hasher,
		Notifier:         notifier,
		Logger:           logger,
		Assistant:        assistant.New(cfg.AssistantProvider, cfg.AssistantWebhookURL, cfg.AssistantWebhookAuth),
		Geo:              geo.New(cfg.GeoProvider, cfg.GeoBaseURL, cfg.GeoUserAgent),
		PaymentsProvider: cfg.PaymentsProvider,
		RefreshTTL:       cfg.JWTRefreshTTL,
		PasswordResetTTL: cfg.PasswordReset,
	})

	server := api.NewServer(api.Config{
		Services:    services,
		Store:       st,
		Tokens:      tokens,
		Logger:      logger,
		CORSOrigins: cfg.CORSAllowedOrigins,
		Pinger:      pool,
	})

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Periodic cleanup of expired sessions.
	go cleanupSessions(ctx, logger, st)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.Addr, "env", string(cfg.Env))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	return httpServer.Shutdown(shutdownCtx)
}

func cleanupSessions(ctx context.Context, logger *slog.Logger, st *store.Store) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			removed, err := st.DeleteExpiredSessions(ctx)
			if err != nil {
				logger.Warn("failed to clean expired sessions", "error", err)
				continue
			}
			if removed > 0 {
				logger.Info("removed expired sessions", "count", removed)
			}
		}
	}
}

// seedAdmin optionally creates a bootstrap administrator account. It is a no-op
// unless SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are provided.
func seedAdmin(ctx context.Context, logger *slog.Logger, st *store.Store, cfg *config.Config) error {
	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPassword == "" {
		return nil
	}
	if _, _, err := st.GetUserWithHash(ctx, auth.NormalizeEmail(cfg.SeedAdminEmail)); err == nil {
		return nil // already exists
	}
	if err := auth.ValidatePassword(cfg.SeedAdminPassword); err != nil {
		return fmt.Errorf("SEED_ADMIN_PASSWORD does not meet the password policy: %w", err)
	}
	hasher := auth.NewPasswordHasher(cfg.BcryptCost)
	hash, err := hasher.Hash(cfg.SeedAdminPassword)
	if err != nil {
		return err
	}
	if _, err := st.CreateUser(ctx, store.CreateUserParams{
		Email:                 auth.NormalizeEmail(cfg.SeedAdminEmail),
		PasswordHash:          hash,
		FullName:              "Administrador",
		Role:                  models.RoleAdmin,
		InstitutionalApproved: true,
	}); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	logger.Info("bootstrap administrator created", "email", cfg.SeedAdminEmail)
	return nil
}
