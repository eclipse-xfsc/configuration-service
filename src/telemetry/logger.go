package telemetry

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global application logger
var Logger *zap.SugaredLogger

// InitializeLogger sets up the global logger with production config
func InitializeLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	config.EncoderConfig.CallerKey = ""
	config.EncoderConfig.StacktraceKey = ""

	logger, _ := config.Build()
	Logger = logger.Sugar()
}
