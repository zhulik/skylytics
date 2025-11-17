package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
	"github.com/mattn/go-isatty"
)

var ErrInvalidLogLevel = errors.New("invalid log level")

func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrInvalidLogLevel, level)
	}
}

func initLogger(level string) error {
	w := os.Stdout

	parsedLevel, err := parseLevel(level)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: parsedLevel,
	}

	var handler slog.Handler
	if isatty.IsTerminal(w.Fd()) {
		handler = devslog.NewHandler(w, &devslog.Options{
			HandlerOptions: opts,
		})
	} else {
		handler = slog.NewJSONHandler(w, nil)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}
