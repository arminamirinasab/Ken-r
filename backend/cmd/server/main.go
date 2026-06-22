// Command server is the Kenâr backend entrypoint. It loads config, the
// bilingual i18n catalogs, wires the use cases to their adapters, builds the
// HTTP router, and serves with graceful shutdown.
//
// The current wiring uses in-memory adapters + an in-process event bus so the
// full API runs locally with zero external infrastructure. The Postgres/Redis/
// MinIO/Pushe/Kavenegar adapters implement the same ports and swap in here as
// they land — see ROADMAP.md.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kenar/backend/internal/adapters/httpapi"
	"github.com/kenar/backend/internal/adapters/memory"
	"github.com/kenar/backend/internal/adapters/sms"
	"github.com/kenar/backend/internal/app/auth"
	"github.com/kenar/backend/internal/app/pairing"
	"github.com/kenar/backend/internal/app/widgets"
	"github.com/kenar/backend/internal/config"
	"github.com/kenar/backend/internal/platform/i18n"
	"github.com/kenar/backend/internal/platform/logger"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	bundle, err := i18n.Load(cfg.I18nDir, cfg.DefaultLocale)
	if err != nil {
		log.Error("i18n load", "err", err)
		os.Exit(1)
	}

	// --- Adapters (in-memory for local/dev; swap to Postgres/Redis later) ---
	users := memory.NewUsers()
	invites := memory.NewInvites()
	pairs := memory.NewPairs()
	devices := memory.NewDevices()
	widgetRepo := memory.NewWidgets()
	otps := memory.NewOTPs()
	sessions := memory.NewSessions(time.Now)
	bus := memory.NewBus()
	smsProvider := sms.NewLogProvider(log) // DEV ONLY: Kavenegar adapter replaces this

	// --- Use cases ---
	authSvc := auth.New(users, otps, sessions, smsProvider, bundle, time.Now)
	pairingSvc := pairing.New(users, invites, pairs, time.Now)
	widgetSvc := widgets.New(widgetRepo, pairs, bus, time.Now)

	router := httpapi.NewRouter(httpapi.Deps{
		Logger:  log,
		I18n:    bundle,
		Auth:    authSvc,
		Pairing: pairingSvc,
		Widgets: widgetSvc,
		Devices: devices,
		Events:  bus,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("kenar backend listening", "addr", cfg.HTTPAddr, "default_locale", cfg.DefaultLocale)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http serve", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
	}
}
