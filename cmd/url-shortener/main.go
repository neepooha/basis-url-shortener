package main

import (
	"context"
	"github.com/neepooha/url_shortener/internal/app"
	"github.com/neepooha/url_shortener/internal/config"
	"github.com/neepooha/url_shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/neepooha/url_shortener/internal/lib/logger/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := app.RunServer(ctx, log, cfg); err != nil {
		log.Error("error to start server", sl.Err(err))
	} else {
		log.Info("server was shutdown")
	}
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
