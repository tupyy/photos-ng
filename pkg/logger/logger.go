package logger

import (
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger initializes and configures a zap logger based on the provided configuration.
// It sets up the appropriate log level and format according to the config settings.
func SetupLogger(cfg *config.Config) *zap.Logger {
	lvl := zapcore.InfoLevel
	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err == nil {
		lvl = level
	}

	loggerCfg := &zap.Config{
		Level:    zap.NewAtomicLevelAt(lvl),
		Encoding: cfg.LogFormat,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "severity",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder, EncodeCaller: zapcore.ShortCallerEncoder},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	plain, err := loggerCfg.Build(zap.AddStacktrace(zap.DPanicLevel))
	if err != nil {
		panic(err)
	}

	return plain
}
