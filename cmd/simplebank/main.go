package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/simplebank/internals/app"
	"github.com/nhassl3/simplebank/internals/config"
	"github.com/nhassl3/simplebank/internals/lib/logger"
)

func main() {
	cfg := config.MustLoadConfig()

	log := logger.MustLoad(cfg.LogType)
	slog.SetDefault(log)

	// If type of logging equal 4 set gin mode of logging to the release mode
	// else continue this stage
	if cfg.LogType == 4 {
		gin.SetMode(gin.ReleaseMode)
	}

	application := app.MustNewApp(
		log,
		cfg.TGP.Secret,
		cfg.ConnectionDBString,
		cfg.Http.Host,
		cfg.Http.Port,
		cfg.TGP.AccessTokenDuration,
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
