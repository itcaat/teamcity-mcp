package logging

import (
	"github.com/itcaat/teamcity-mcp/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new structured logger
func New(cfg config.LoggingConfig) (*zap.SugaredLogger, error) {
	var zapConfig zap.Config

	if cfg.Format == "console" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Add correlation ID and trace context to logs
	zapConfig.InitialFields = map[string]interface{}{
		"service": "teamcity-mcp",
	}

	logger, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

// WithRequestID adds a request ID to the logger context
func WithRequestID(logger *zap.SugaredLogger, requestID string) *zap.SugaredLogger {
	return logger.With("request_id", requestID)
}

// WithTraceID adds OpenTelemetry trace information to the logger
func WithTraceID(logger *zap.SugaredLogger, traceID, spanID string) *zap.SugaredLogger {
	return logger.With("trace_id", traceID, "span_id", spanID)
}
