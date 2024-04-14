package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	ssogrpc "url_shortener/internal/clients/sso/grpc"
	"url_shortener/internal/config"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/lib/migrator"
	"url_shortener/internal/storage/postgres"
	admDel "url_shortener/internal/transport/handlers/admins/delete"
	admSet "url_shortener/internal/transport/handlers/admins/set"
	urlDel "url_shortener/internal/transport/handlers/url/delete"
	urlRed "url_shortener/internal/transport/handlers/url/redirect"
	urlSave "url_shortener/internal/transport/handlers/url/save"
	"url_shortener/internal/transport/middleware/auth"
	"url_shortener/internal/transport/middleware/isadmin"
	mwLogger "url_shortener/internal/transport/middleware/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
)

func RunServer(ctx context.Context, log *slog.Logger, cfg *config.Config) error {
	const op = "internal.app.RunServer"
	log.With(slog.String("op", op))

	// init ssoServer
	log.Info("init ssoClinet", slog.String("env", cfg.Env))
	log.Debug("creddentials sso", slog.String("address", cfg.Clients.SSO.Address))
	ssoClient, err := ssogrpc.New(
		context.Background(),
		log, cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init ssoClient", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("ssoClient was init")

	// init postgresql storage
	storage, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer storage.CloseStorage()

	// start migration
	err = migrator.Migrate(cfg)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug("no migrations to apply")
		} else {
			panic(err)
		}
	}
	log.Debug("migrations applied successfully")

	// init router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// url router
	router.Route("/url", func(r chi.Router) {
		r.Use(auth.New(log, cfg.AppSecret))
		r.Post("/", urlSave.New(log, storage))
	})
	router.Route("/url/{alias}", func(r chi.Router) {
		r.Use(auth.New(log, cfg.AppSecret))
		r.Use(isadmin.New(log, ssoClient))
		r.Delete("/", urlDel.New(log, storage))
	})
	router.Get("/{alias}", urlRed.New(log, storage))

	// user router
	router.Route("/user", func(r chi.Router) {
		r.Post("/", admSet.New(log, ssoClient))
		r.Delete("/", admDel.New(log, ssoClient))
	})

	// start server
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server")
			os.Exit(1)
		}
	}()
	log.Info("url shortener is running", slog.String("addresses", srv.Addr))

	// wait for gracefully shutdown
	<-ctx.Done()
	log.Info("shutting down server gracefully")
	shutDownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutDownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	<-shutDownCtx.Done()
	return nil
}
