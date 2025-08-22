package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DebugLogger provides debug-only logging for business services
// IMPORTANT: Never logs errors (that's handled by handlers)
type DebugLogger struct {
	logger  *zap.Logger
	service string
}

// NewDebugLogger creates a new debug logger for a specific service
func NewDebugLogger(service string) *DebugLogger {
	return &DebugLogger{
		logger:  zap.L().Named(service),
		service: service,
	}
}

// WithContext returns a new DebugLogger with request context
func (l *DebugLogger) WithContext(ctx context.Context) *DebugLogger {
	// Extract request ID if available
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return &DebugLogger{
			logger:  l.logger.With(zap.String("request_id", requestID)),
			service: l.service,
		}
	}
	return l
}

// StartOperation begins operation tracing and returns a builder
func (l *DebugLogger) StartOperation(operation string) *OperationBuilder {
	return &OperationBuilder{
		operation: operation,
		params:    make(map[string]any),
		logger:    l,
	}
}

// StartOperationWithParams begins operation tracing with map (deprecated, use StartOperation().WithParam())
func (l *DebugLogger) StartOperationWithParams(operation string, params map[string]any) *OperationTracer {
	start := time.Now()

	l.logger.Debug("Operation started",
		zap.String("operation", operation),
		zap.Any("params", params),
	)

	return &OperationTracer{
		DebugLogger: l,
		operation:   operation,
		startTime:   start,
		params:      params,
	}
}

// OperationBuilder builds operation parameters fluently
type OperationBuilder struct {
	operation string
	params    map[string]any
	logger    *DebugLogger
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

	b.logger.logger.Debug("Operation started",
		zap.String("operation", b.operation),
		zap.Any("params", b.params),
	)

	return &OperationTracer{
		DebugLogger: b.logger,
		operation:   b.operation,
		startTime:   start,
		params:      b.params,
	}
}

// OperationTracer tracks the progress of a business operation
type OperationTracer struct {
	*DebugLogger
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
	ot.logger.Debug("Operation step",
		zap.String("operation", ot.operation),
		zap.String("step", step),
		zap.Float64("elapsed_ms", float64(time.Since(ot.startTime).Nanoseconds())/1e6),
		zap.Any("data", data),
	)
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
	ot.logger.Debug("Operation completed",
		zap.String("operation", ot.operation),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		zap.Any("result", result),
	)
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
	b.tracer.logger.Debug("Operation step",
		zap.String("operation", b.tracer.operation),
		zap.String("step", b.step),
		zap.Float64("elapsed_ms", float64(time.Since(b.tracer.startTime).Nanoseconds())/1e6),
		zap.Any("data", b.data),
	)
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
	b.tracer.logger.Debug("Operation completed",
		zap.String("operation", b.tracer.operation),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		zap.Any("result", b.result),
	)
}

// Performance logs a performance metric during the operation
func (ot *OperationTracer) Performance(metric string, value any) {
	ot.logger.Debug("Performance metric",
		zap.String("operation", ot.operation),
		zap.String("metric", metric),
		zap.Any("value", value),
		zap.Float64("elapsed_ms", float64(time.Since(ot.startTime).Nanoseconds())/1e6),
	)
}

// Convenience methods for common debug scenarios

// DatabaseQuery logs database query performance information
func (l *DebugLogger) DatabaseQuery(operation string, filters int, duration time.Duration, found bool) {
	l.logger.Debug("Database query",
		zap.String("operation", operation),
		zap.Int("filters", filters),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		zap.Bool("found", found),
	)
}

// FileOperation logs file system operation details
func (l *DebugLogger) FileOperation(operation, filepath string, size int64, duration time.Duration) {
	l.logger.Debug("File operation",
		zap.String("operation", operation),
		zap.String("filepath", filepath),
		zap.Int64("size", size),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
	)
}

// BusinessLogic creates a business logic data builder
func (l *DebugLogger) BusinessLogic(description string) *DataBuilder {
	return &DataBuilder{
		logger:      l,
		logType:     "Business logic",
		description: description,
		data:        make(map[string]any),
	}
}

// BusinessLogicWithData logs business logic with map data (deprecated, use BusinessLogic().WithParam())
func (l *DebugLogger) BusinessLogicWithData(description string, data map[string]any) {
	l.logger.Debug("Business logic",
		zap.String("description", description),
		zap.Any("data", data),
	)
}

// Transaction creates a transaction data builder
func (l *DebugLogger) Transaction(action string) *DataBuilder {
	return &DataBuilder{
		logger:  l,
		logType: "Transaction",
		action:  action,
		data:    make(map[string]any),
	}
}

// TransactionWithData logs transaction with map data (deprecated, use Transaction().WithParam())
func (l *DebugLogger) TransactionWithData(action string, data map[string]any) {
	l.logger.Debug("Transaction",
		zap.String("action", action),
		zap.Any("data", data),
	)
}

// Processing creates a processing data builder
func (l *DebugLogger) Processing(stage, filename string) *DataBuilder {
	return &DataBuilder{
		logger:   l,
		logType:  "Processing",
		stage:    stage,
		filename: filename,
		data:     make(map[string]any),
	}
}

// ProcessingWithData logs processing with map data (deprecated, use Processing().WithParam())
func (l *DebugLogger) ProcessingWithData(stage string, filename string, data map[string]any) {
	l.logger.Debug("Processing",
		zap.String("stage", stage),
		zap.String("filename", filename),
		zap.Any("data", data),
	)
}

// DataBuilder builds logging data fluently for convenience methods
type DataBuilder struct {
	logger      *DebugLogger
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
	switch b.logType {
	case "Business logic":
		b.logger.logger.Debug("Business logic",
			zap.String("description", b.description),
			zap.Any("data", b.data),
		)
	case "Transaction":
		b.logger.logger.Debug("Transaction",
			zap.String("action", b.action),
			zap.Any("data", b.data),
		)
	case "Processing":
		b.logger.logger.Debug("Processing",
			zap.String("stage", b.stage),
			zap.String("filename", b.filename),
			zap.Any("data", b.data),
		)
	}
}
