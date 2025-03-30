package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"vk-test-assignment-mattermost-polls/pkg/config"
)

func Setup(cfg config.LoggerConfig) {
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout

	if cfg.Output == "file" && cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			output = file
		}
	}

	var writer = output
	if cfg.Format == "pretty" {
		writer = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	logger := zerolog.New(writer).With().Timestamp()

	if cfg.WithCaller {
		logger = logger.Caller()
	}

	log.Logger = logger.Logger()

	log.Info().
		Str("level", level.String()).
		Str("format", cfg.Format).
		Str("output", cfg.Output).
		Bool("with_caller", cfg.WithCaller).
		Msg("Logger initialized")
}
