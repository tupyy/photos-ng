package logger

import (
	"context"
	"time"

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

// StructuredLogger provides structured logging for business services at specified level
type StructuredLogger struct {
	logger  *zap.Logger
	service string
	level   zapcore.Level
}

// NewStructuredLogger creates a new structured logger for a specific service at the given level
func NewStructuredLogger(service string, level zapcore.Level) *StructuredLogger {
	return &StructuredLogger{
		logger:  zap.L().Named(service),
		service: service,
		level:   level,
	}
}

// NewDebugLogger creates a new debug-level structured logger for a specific service
func NewDebugLogger(service string) *StructuredLogger {
	return NewStructuredLogger(service, zapcore.DebugLevel)
}

// NewInfoLogger creates a new info-level structured logger for a specific service
func NewInfoLogger(service string) *StructuredLogger {
	return NewStructuredLogger(service, zapcore.InfoLevel)
}

// NewWarnLogger creates a new warn-level structured logger for a specific service
func NewWarnLogger(service string) *StructuredLogger {
	return NewStructuredLogger(service, zapcore.WarnLevel)
}

// NewErrorLogger creates a new error-level structured logger for a specific service
func NewErrorLogger(service string) *StructuredLogger {
	return NewStructuredLogger(service, zapcore.ErrorLevel)
}

// DebugLogger is an alias for backward compatibility
type DebugLogger = StructuredLogger

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

// StartOperation begins operation tracing and returns a builder
func (l *StructuredLogger) StartOperation(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		params:    make(map[string]any),
		logger:    l,
	}
}

// StartOperationWithParams begins operation tracing with map (deprecated, use StartOperation().WithParam())
func (l *StructuredLogger) StartOperationWithParams(operation string, params map[string]any) *OperationTracer {
	start := time.Now()

	logFunc := l.getLogFunc()
	logFunc("Operation started", zap.String("operation", operation), zap.Any("params", params))

	return &OperationTracer{
		StructuredLogger: l,
		operation:        operation,
		startTime:        start,
		params:           params,
	}
}

// OperationBuilder builds operation parameters fluently
type OperationBuilder struct {
	operation string
	params    map[string]any
	logger    *StructuredLogger
}

// WithParam adds a generic parameter
func (b *OperationBuilder) WithParam(key string, value any) *OperationBuilder {
	b.params[key] = value
	return b
}

// WithString adds a string parameter
func (b *OperationBuilder) WithString(key, value string) *OperationBuilder {
	return b.WithParam(key, value)
}

// WithInt adds an int parameter
func (b *OperationBuilder) WithInt(key string, value int) *OperationBuilder {
	return b.WithParam(key, value)
}

// WithBool adds a bool parameter
func (b *OperationBuilder) WithBool(key string, value bool) *OperationBuilder {
	return b.WithParam(key, value)
}

// WithStringPtr adds a string pointer parameter (nil-safe)
func (b *OperationBuilder) WithStringPtr(key string, value *string) *OperationBuilder {
	if value != nil {
		return b.WithParam(key, *value)
	}
	return b.WithParam(key, nil)
}

// WithIntPtr adds an int pointer parameter (nil-safe)
func (b *OperationBuilder) WithIntPtr(key string, value *int) *OperationBuilder {
	if value != nil {
		return b.WithParam(key, *value)
	}
	return b.WithParam(key, nil)
}

// Build creates and starts the operation tracer
func (b *OperationBuilder) Build() *OperationTracer {
	start := time.Now()

	logFunc := b.logger.getLogFunc()
	logFunc("Operation started", zap.String("operation", b.operation), zap.Any("params", b.params))

	return &OperationTracer{
		StructuredLogger: b.logger,
		operation:        b.operation,
		startTime:        start,
		params:           b.params,
	}
}

// OperationTracer tracks the progress of a business operation
type OperationTracer struct {
	*StructuredLogger
	operation string
	startTime time.Time
	params    map[string]any
}

// Step creates a step builder
func (ot *OperationTracer) Step(step string) *StepBuilder {
	return &StepBuilder{
		tracer: ot,
		step:   step,
		data:   make(map[string]any),
	}
}

// StepWithData logs a step with map data (deprecated, use Step().WithParam())
func (ot *OperationTracer) StepWithData(step string, data map[string]any) {
	logFunc := ot.getLogFunc()
	logFunc("Operation step", zap.String("operation", ot.operation), zap.String("step", step), zap.Float64("elapsed_ms", float64(time.Since(ot.startTime).Nanoseconds())/1e6), zap.Any("data", data))
}

// Success creates a result builder
func (ot *OperationTracer) Success() *ResultBuilder {
	return &ResultBuilder{
		tracer: ot,
		result: make(map[string]any),
	}
}

// SuccessWithResult logs success with map result (deprecated, use Success().WithParam())
func (ot *OperationTracer) SuccessWithResult(result map[string]any) {
	duration := time.Since(ot.startTime)
	logFunc := ot.getLogFunc()
	logFunc("Operation completed", zap.String("operation", ot.operation), zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6), zap.Any("result", result))
}

// StepBuilder builds step data fluently
type StepBuilder struct {
	tracer *OperationTracer
	step   string
	data   map[string]any
}

// WithParam adds a generic parameter to the step
func (b *StepBuilder) WithParam(key string, value any) *StepBuilder {
	b.data[key] = value
	return b
}

// WithString adds a string parameter to the step
func (b *StepBuilder) WithString(key, value string) *StepBuilder {
	return b.WithParam(key, value)
}

// WithInt adds an int parameter to the step
func (b *StepBuilder) WithInt(key string, value int) *StepBuilder {
	return b.WithParam(key, value)
}

// WithBool adds a bool parameter to the step
func (b *StepBuilder) WithBool(key string, value bool) *StepBuilder {
	return b.WithParam(key, value)
}

// WithStringPtr adds a string pointer parameter to the step (nil-safe)
func (b *StepBuilder) WithStringPtr(key string, value *string) *StepBuilder {
	if value != nil {
		return b.WithParam(key, *value)
	}
	return b.WithParam(key, nil)
}

// WithDurationMs adds a duration parameter converted to milliseconds
func (b *StepBuilder) WithDurationMs(key string, value time.Duration) *StepBuilder {
	return b.WithParam(key, float64(value.Nanoseconds())/1e6)
}

// Log executes the step logging
func (b *StepBuilder) Log() {
	logFunc := b.tracer.getLogFunc()
	logFunc("Operation step", zap.String("operation", b.tracer.operation), zap.String("step", b.step), zap.Float64("elapsed_ms", float64(time.Since(b.tracer.startTime).Nanoseconds())/1e6), zap.Any("data", b.data))
}

// ResultBuilder builds result data fluently
type ResultBuilder struct {
	tracer *OperationTracer
	result map[string]any
}

// WithParam adds a generic parameter to the result
func (b *ResultBuilder) WithParam(key string, value any) *ResultBuilder {
	b.result[key] = value
	return b
}

// WithString adds a string parameter to the result
func (b *ResultBuilder) WithString(key, value string) *ResultBuilder {
	return b.WithParam(key, value)
}

// WithInt adds an int parameter to the result
func (b *ResultBuilder) WithInt(key string, value int) *ResultBuilder {
	return b.WithParam(key, value)
}

// WithBool adds a bool parameter to the result
func (b *ResultBuilder) WithBool(key string, value bool) *ResultBuilder {
	return b.WithParam(key, value)
}

// WithStringPtr adds a string pointer parameter to the result (nil-safe)
func (b *ResultBuilder) WithStringPtr(key string, value *string) *ResultBuilder {
	if value != nil {
		return b.WithParam(key, *value)
	}
	return b.WithParam(key, nil)
}

// WithDurationMs adds a duration parameter converted to milliseconds
func (b *ResultBuilder) WithDurationMs(key string, value time.Duration) *ResultBuilder {
	return b.WithParam(key, float64(value.Nanoseconds())/1e6)
}

// Log executes the success logging
func (b *ResultBuilder) Log() {
	duration := time.Since(b.tracer.startTime)
	logFunc := b.tracer.getLogFunc()
	logFunc("Operation completed", zap.String("operation", b.tracer.operation), zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6), zap.Any("result", b.result))
}

// Performance logs a performance metric during the operation
func (ot *OperationTracer) Performance(metric string, value any) {
	logFunc := ot.getLogFunc()
	logFunc("Performance metric", zap.String("operation", ot.operation), zap.String("metric", metric), zap.Any("value", value), zap.Float64("elapsed_ms", float64(time.Since(ot.startTime).Nanoseconds())/1e6))
}

// Convenience methods for common debug scenarios

// DatabaseQuery logs database query performance information
func (l *StructuredLogger) DatabaseQuery(operation string, filters int, duration time.Duration, found bool) {
	logFunc := l.getLogFunc()
	logFunc("Database query", zap.String("operation", operation), zap.Int("filters", filters), zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6), zap.Bool("found", found))
}

// FileOperation logs file system operation details
func (l *StructuredLogger) FileOperation(operation, filepath string, size int64, duration time.Duration) {
	logFunc := l.getLogFunc()
	logFunc("File operation", zap.String("operation", operation), zap.String("filepath", filepath), zap.Int64("size", size), zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6))
}

// BusinessLogic creates a business logic data builder
func (l *StructuredLogger) BusinessLogic(description string) *DataBuilder {
	return &DataBuilder{
		logger:      l,
		logType:     "Business logic",
		description: description,
		data:        make(map[string]any),
	}
}

// BusinessLogicWithData logs business logic with map data (deprecated, use BusinessLogic().WithParam())
func (l *StructuredLogger) BusinessLogicWithData(description string, data map[string]any) {
	logFunc := l.getLogFunc()
	logFunc("Business logic", zap.String("description", description), zap.Any("data", data))
}

// Transaction creates a transaction data builder
func (l *StructuredLogger) Transaction(action string) *DataBuilder {
	return &DataBuilder{
		logger:  l,
		logType: "Transaction",
		action:  action,
		data:    make(map[string]any),
	}
}

// TransactionWithData logs transaction with map data (deprecated, use Transaction().WithParam())
func (l *StructuredLogger) TransactionWithData(action string, data map[string]any) {
	logFunc := l.getLogFunc()
	logFunc("Transaction", zap.String("action", action), zap.Any("data", data))
}

// Processing creates a processing data builder
func (l *StructuredLogger) Processing(stage, filename string) *DataBuilder {
	return &DataBuilder{
		logger:   l,
		logType:  "Processing",
		stage:    stage,
		filename: filename,
		data:     make(map[string]any),
	}
}

// ProcessingWithData logs processing with map data (deprecated, use Processing().WithParam())
func (l *StructuredLogger) ProcessingWithData(stage string, filename string, data map[string]any) {
	logFunc := l.getLogFunc()
	logFunc("Processing", zap.String("stage", stage), zap.String("filename", filename), zap.Any("data", data))
}

// DataBuilder builds logging data fluently for convenience methods
type DataBuilder struct {
	logger      *StructuredLogger
	logType     string
	description string
	action      string
	stage       string
	filename    string
	data        map[string]any
}

// WithParam adds a generic parameter to the data
func (b *DataBuilder) WithParam(key string, value any) *DataBuilder {
	b.data[key] = value
	return b
}

// WithString adds a string parameter to the data
func (b *DataBuilder) WithString(key, value string) *DataBuilder {
	return b.WithParam(key, value)
}

// WithInt adds an int parameter to the data
func (b *DataBuilder) WithInt(key string, value int) *DataBuilder {
	return b.WithParam(key, value)
}

// WithBool adds a bool parameter to the data
func (b *DataBuilder) WithBool(key string, value bool) *DataBuilder {
	return b.WithParam(key, value)
}

// WithStringPtr adds a string pointer parameter to the data (nil-safe)
func (b *DataBuilder) WithStringPtr(key string, value *string) *DataBuilder {
	if value != nil {
		return b.WithParam(key, *value)
	}
	return b.WithParam(key, nil)
}

// WithDurationMs adds a duration parameter converted to milliseconds
func (b *DataBuilder) WithDurationMs(key string, value time.Duration) *DataBuilder {
	return b.WithParam(key, float64(value.Nanoseconds())/1e6)
}

// Log executes the appropriate logging based on type
func (b *DataBuilder) Log() {
	logFunc := b.logger.getLogFunc()
	switch b.logType {
	case "Business logic":
		logFunc("Business logic", zap.String("description", b.description), zap.Any("data", b.data))
	case "Transaction":
		logFunc("Transaction", zap.String("action", b.action), zap.Any("data", b.data))
	case "Processing":
		logFunc("Processing", zap.String("stage", b.stage), zap.String("filename", b.filename), zap.Any("data", b.data))
	}
}
