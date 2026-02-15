package log

import (
	"context"

	"github.com/rs/zerolog"
)

// GooseLogger adapts zerolog to goose's Logger interface
type GooseLogger struct {
	logger *zerolog.Logger
}

func (g *GooseLogger) Fatalf(format string, v ...interface{}) {
	g.logger.Fatal().Msgf(format, v...)
}

func (g *GooseLogger) Printf(format string, v ...interface{}) {
	g.logger.Info().Msgf(format, v...)
}

func NewGooseLoggerFromCtx(ctx context.Context) *GooseLogger {
	return &GooseLogger{
		logger: FromCtx(ctx),
	}
}
