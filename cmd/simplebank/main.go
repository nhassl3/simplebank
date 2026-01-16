package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/nhassl3/simplebank/internals/app"
	"github.com/nhassl3/simplebank/internals/config"
	"github.com/nhassl3/simplebank/internals/lib/logger"
)

func main() {
	cfg := config.MustLoadConfig()

	log := logger.MustLoad(cfg.EnvDefault)
	slog.SetDefault(log)

	application := app.MustNewApp(
		log,
		cfg.JWT.Secret,
		cfg.ConnectionDBString,
		cfg.Http.Host,
		cfg.Http.Port,
		cfg.JWT.AccessTokenDuration,
	)

	log.Info("starting application")
	go application.MustStart()
	log.Info("application started", slog.String("Host", cfg.Http.Host), slog.Int("Port", cfg.Http.Port))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	s := <-sig

	application.Stop()

	log.Info("received signal", slog.String("s", s.String()))
}
