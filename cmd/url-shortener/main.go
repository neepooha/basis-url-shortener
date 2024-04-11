package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	ssogrpc "url_shortener/internal/clients/sso/grpc"
	"url_shortener/internal/config"
	"url_shortener/internal/lib/logger/handlers/slogpretty"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage/postgres"

	// "url_shortener/internal/storage/sqlite"
	urlDel "url_shortener/internal/transport/handlers/url/delete"
	urlRed"url_shortener/internal/transport/handlers/url/redirect"
	urlSave "url_shortener/internal/transport/handlers/url/save"
	admSet"url_shortener/internal/transport/handlers/admins/set"
	admDel "url_shortener/internal/transport/handlers/admins/delete"
	"url_shortener/internal/transport/middleware/auth"
	mwLogger "url_shortener/internal/transport/middleware/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := setupLogger(cfg.Env)
	log.Info("starting url shortener", slog.String("env", cfg.Env))
	log.Debug("creddentials url-shortener", slog.String("address", cfg.Address))

	// connect to ssoServer
	log.Info("init ssoServer", slog.String("env", cfg.Env))
	log.Debug("creddentials sso", slog.String("address", cfg.Clients.SSO.Address))

	ssoClient, err := ssogrpc.New(
		context.Background(),
		log, cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init ssoClient", sl.Err(err))
		os.Exit(1)
	}
	log.Info("ssoClient was init")

	// init postgresql storage
	storage, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	/* // init sqlite storage
	storage, err := sqlite.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	} */

	// init router
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// url router
	router.Route("/url", func(r chi.Router) {
		r.Use(auth.New(log, cfg.AppSecret, ssoClient))

		r.Post("/", urlSave.New(log, storage))
		r.Delete("/{alias}", urlDel.New(log, storage))
	})
	router.Get("/{alias}", urlRed.New(log, storage))

	/// user router
	router.Route("/user", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url_shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", admSet.New(log, ssoClient))
		r.Delete("/", admDel.New(log, ssoClient))
	})

	// start server
	log.Info("starting server", slog.String("addresses", cfg.Address))
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
	log.Error("server was stoped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
