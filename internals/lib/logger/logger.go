package logger

import (
	"log/slog"
	"os"

	prettyslogger "github.com/nhassl3/simplebank/internals/lib/logger/handler/prettysloger"
)

const (
	_ = iota
	EnvLocal
	EnvTest
	EnvDev
	EnvProd
)

func MustLoad(envLevel int) *slog.Logger {
	var level slog.Level
	switch envLevel {
	case EnvLocal:
		level = slog.LevelDebug
	case EnvTest:
		level = slog.LevelDebug
	case EnvDev:
		level = slog.LevelWarn
	case EnvProd:
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}

	opts := prettyslogger.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: level,
		},
	}

	return slog.New(opts.NewPrettyLogger(os.Stdout))
}
