package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/requestid"
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
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	plain, err := loggerCfg.Build(zap.AddStacktrace(zap.DPanicLevel))
	if err != nil {
		panic(err)
	}

	return plain
}

// StructuredLogger provides structured logging for business services at specified level
type StructuredLogger struct {
	logger  *zap.Logger
	service string
	level   zapcore.Level
}

// New creates a new structured logger for a specific service
// The default level is InfoLevel, but you can log at any level using Debug(), Info(), Warn(), Error() methods
func New(service string) *StructuredLogger {
	return &StructuredLogger{
		logger:  zap.L().Named(service),
		service: service,
		level:   zapcore.InfoLevel, // Default level
	}
}

// getLogFunc returns the appropriate logging function based on the configured level
func (l *StructuredLogger) getLogFunc() func(msg string, fields ...zap.Field) {
	switch l.level {
	case zapcore.DebugLevel:
		return l.logger.Debug
	case zapcore.InfoLevel:
		return l.logger.Info
	case zapcore.WarnLevel:
		return l.logger.Warn
	case zapcore.ErrorLevel:
		return l.logger.Error
	default:
		return l.logger.Debug
	}
}

// WithContext returns a new StructuredLogger with request context
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	// Extract request ID if available
	if requestID := requestid.FromContext(ctx); requestID != "" {
		return &StructuredLogger{
			logger:  l.logger.With(zap.String("request_id", requestID)),
			service: l.service,
			level:   l.level,
		}
	}
	return l
}

// Operation begins operation tracing and returns a builder at the logger's default level
func (l *StructuredLogger) Operation(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		fields:    make([]zap.Field, 0),
		logger:    l,
		level:     l.level,
	}
}

// Debug begins operation tracing at debug level
func (l *StructuredLogger) Debug(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		fields:    make([]zap.Field, 0),
		logger:    l,
		level:     zapcore.DebugLevel,
	}
}

// Info begins operation tracing at info level
func (l *StructuredLogger) Info(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		fields:    make([]zap.Field, 0),
		logger:    l,
		level:     zapcore.InfoLevel,
	}
}

// Warn begins operation tracing at warn level
func (l *StructuredLogger) Warn(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		fields:    make([]zap.Field, 0),
		logger:    l,
		level:     zapcore.WarnLevel,
	}
}

// Error begins operation tracing at error level
func (l *StructuredLogger) Error(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		fields:    make([]zap.Field, 0),
		logger:    l,
		level:     zapcore.ErrorLevel,
	}
}

// OperationBuilder builds operation parameters fluently
type OperationBuilder struct {
	operation string
	fields    []zap.Field
	logger    *StructuredLogger
	level     zapcore.Level
}

// WithParam adds a generic parameter
func (b *OperationBuilder) WithParam(key string, value any) *OperationBuilder {
	b.fields = append(b.fields, zap.Any(key, value))
	return b
}

// WithString adds a string parameter
func (b *OperationBuilder) WithString(key, value string) *OperationBuilder {
	b.fields = append(b.fields, zap.String(key, value))
	return b
}

// WithInt adds an int parameter
func (b *OperationBuilder) WithInt(key string, value int) *OperationBuilder {
	b.fields = append(b.fields, zap.Int(key, value))
	return b
}

// WithBool adds a bool parameter
func (b *OperationBuilder) WithBool(key string, value bool) *OperationBuilder {
	b.fields = append(b.fields, zap.Bool(key, value))
	return b
}

// WithStringPtr adds a string pointer parameter (nil-safe)
func (b *OperationBuilder) WithStringPtr(key string, value *string) *OperationBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.String(key, *value))
	} else {
		b.fields = append(b.fields, zap.Any(key, nil))
	}
	return b
}

// WithIntPtr adds an int pointer parameter (nil-safe)
func (b *OperationBuilder) WithIntPtr(key string, value *int) *OperationBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.Int(key, *value))
	} else {
		b.fields = append(b.fields, zap.Any(key, nil))
	}
	return b
}

// WithRequestBody adds a request body parameter
func (b *OperationBuilder) WithRequestBody(key string, value any) *OperationBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.Any(key, value))
	}
	return b
}

// Build creates and starts the operation tracer
func (b *OperationBuilder) Build() *OperationTracer {
	logFunc := getLogFuncForLevel(b.logger.logger, b.level)
	fields := append([]zap.Field{}, b.fields...)
	logFunc("operation started", fields...)

	return &OperationTracer{
		StructuredLogger: b.logger,
		operation:        b.operation,
		fields:           b.fields,
		level:            b.level,
	}
}

// OperationTracer tracks the progress of a business operation
type OperationTracer struct {
	*StructuredLogger
	operation string
	fields    []zap.Field
	level     zapcore.Level
}

// getLogFuncForLevel returns the appropriate logging function for the given level
func getLogFuncForLevel(logger *zap.Logger, level zapcore.Level) func(msg string, fields ...zap.Field) {
	switch level {
	case zapcore.DebugLevel:
		return logger.Debug
	case zapcore.InfoLevel:
		return logger.Info
	case zapcore.WarnLevel:
		return logger.Warn
	case zapcore.ErrorLevel:
		return logger.Error
	default:
		return logger.Debug
	}
}

// Step creates a step builder
func (ot *OperationTracer) Step(step string) *StepBuilder {
	return &StepBuilder{
		tracer: ot,
		step:   step,
		fields: make([]zap.Field, 0),
	}
}

// Success creates a result builder
func (ot *OperationTracer) Success() *ResultBuilder {
	return &ResultBuilder{
		tracer:  ot,
		fields:  make([]zap.Field, 0),
		isError: false,
	}
}

// Error creates an error result builder that logs at error level
func (ot *OperationTracer) Error(err error) *ResultBuilder {
	return &ResultBuilder{
		tracer:  ot,
		fields:  []zap.Field{zap.String("error", err.Error())},
		isError: true,
	}
}

// StepBuilder builds step data fluently
type StepBuilder struct {
	tracer *OperationTracer
	step   string
	fields []zap.Field
}

// WithParam adds a generic parameter to the step
func (b *StepBuilder) WithParam(key string, value any) *StepBuilder {
	b.fields = append(b.fields, zap.Any(key, value))
	return b
}

// WithString adds a string parameter to the step
func (b *StepBuilder) WithString(key, value string) *StepBuilder {
	b.fields = append(b.fields, zap.String(key, value))
	return b
}

// WithInt adds an int parameter to the step
func (b *StepBuilder) WithInt(key string, value int) *StepBuilder {
	b.fields = append(b.fields, zap.Int(key, value))
	return b
}

// WithBool adds a bool parameter to the step
func (b *StepBuilder) WithBool(key string, value bool) *StepBuilder {
	b.fields = append(b.fields, zap.Bool(key, value))
	return b
}

// WithStringPtr adds a string pointer parameter to the step (nil-safe)
func (b *StepBuilder) WithStringPtr(key string, value *string) *StepBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.String(key, *value))
	} else {
		b.fields = append(b.fields, zap.Any(key, nil))
	}
	return b
}

// WithIntPtr adds an int pointer parameter to the step (nil-safe)
func (b *StepBuilder) WithIntPtr(key string, value *int) *StepBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.Int(key, *value))
	} else {
		b.fields = append(b.fields, zap.Any(key, nil))
	}
	return b
}

// Log executes the step logging
func (b *StepBuilder) Log() {
	logFunc := getLogFuncForLevel(b.tracer.logger, b.tracer.level)
	fields := []zap.Field{
		zap.String("operation", b.tracer.operation),
		zap.String("step", b.step),
	}
	fields = append(fields, b.fields...)
	logFunc("operation step", fields...)
}

// ResultBuilder builds result data fluently
type ResultBuilder struct {
	tracer  *OperationTracer
	fields  []zap.Field
	isError bool
}

// WithParam adds a generic parameter to the result
func (b *ResultBuilder) WithParam(key string, value any) *ResultBuilder {
	b.fields = append(b.fields, zap.Any(key, value))
	return b
}

// WithString adds a string parameter to the result
func (b *ResultBuilder) WithString(key, value string) *ResultBuilder {
	b.fields = append(b.fields, zap.String(key, value))
	return b
}

// WithInt adds an int parameter to the result
func (b *ResultBuilder) WithInt(key string, value int) *ResultBuilder {
	b.fields = append(b.fields, zap.Int(key, value))
	return b
}

// WithBool adds a bool parameter to the result
func (b *ResultBuilder) WithBool(key string, value bool) *ResultBuilder {
	b.fields = append(b.fields, zap.Bool(key, value))
	return b
}

// WithStringPtr adds a string pointer parameter to the result (nil-safe)
func (b *ResultBuilder) WithStringPtr(key string, value *string) *ResultBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.String(key, *value))
	} else {
		b.fields = append(b.fields, zap.Any(key, nil))
	}
	return b
}

// WithIntPtr adds an int pointer parameter to the result (nil-safe)
func (b *ResultBuilder) WithIntPtr(key string, value *int) *ResultBuilder {
	if value != nil {
		b.fields = append(b.fields, zap.Int(key, *value))
	} else {
		b.fields = append(b.fields, zap.Any(key, nil))
	}
	return b
}

// WithError adds the error to the log (typically used with Failed())
func (b *ResultBuilder) WithError(err error) *ResultBuilder {
	b.fields = append(b.fields, zap.String("error", err.Error()))
	return b
}

// Log executes the result logging
func (b *ResultBuilder) Log() {
	fields := []zap.Field{
		zap.String("operation", b.tracer.operation),
	}
	fields = append(fields, b.fields...)

	if b.isError {
		b.tracer.logger.Error("operation failed", fields...)
	} else {
		logFunc := getLogFuncForLevel(b.tracer.logger, b.tracer.level)
		logFunc("operation completed", fields...)
	}
}
