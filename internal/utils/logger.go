package utils

import (
	"os"

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/jivvon/node-label-controller/internal/constants"
)

// GetLogLevelFromEnv reads log level from environment variable
func GetLogLevelFromEnv() constants.LogLevel {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		return constants.LogLevel(logLevel)
	}
	return constants.LogLevelInfo // default to INFO
}

// GetZapLevel converts LogLevel to zapcore.Level
func GetZapLevel(logLevel constants.LogLevel) zapcore.Level {
	switch logLevel {
	case constants.LogLevelDebug:
		return zapcore.DebugLevel
	case constants.LogLevelInfo:
		return zapcore.InfoLevel
	case constants.LogLevelWarning:
		return zapcore.WarnLevel
	case constants.LogLevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel // default to INFO
	}
}

// SetupLogger configures the logger with the given options
func SetupLogger(opts *zap.Options) {
	logLevel := GetLogLevelFromEnv()

	if !logLevel.IsValid() {
		// Log error but continue with default level
		zap.New(zap.UseFlagOptions(opts)).Error(nil, "Invalid LOG_LEVEL",
			"value", logLevel.String(),
			"supported", constants.GetSupportedLevels())
		return
	}

	opts.Level = GetZapLevel(logLevel)
}
