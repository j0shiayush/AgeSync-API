// Package logger provides a thin wrapper around uber-go/zap to initialise
// a production-quality structured logger with caller info and ISO-8601 timestamps.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a *zap.Logger configured for the given environment.
// In "development" mode the logger uses a human-readable console encoder;
// in all other environments it outputs structured JSON suitable for log aggregators.
func New(env string) (*zap.Logger, error) {
	var cfg zap.Config

	if env == "development" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, fmt.Errorf("logger: failed to build zap logger: %w", err)
	}

	return logger, nil
}