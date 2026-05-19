package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LoggerKey          = "logger"
	loggerRequestIDKey = "x-request-id"
	loggerTraceIDKey   = "x-trace-id"
	LoggerEnvKey       = "env"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
	Sync() error
}

type L struct {
	z *zap.Logger
}

func NewLogger(env string) (Logger, error) {
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(
		"logs/app.json",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	level := zap.InfoLevel

	if env == "debug" {
		level = zap.DebugLevel
	}

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileWriter := zapcore.AddSync(logFile)
	consoleWriter := zapcore.AddSync(os.Stdout)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)

	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	loggerStruct := &L{logger}
	return loggerStruct, nil
}

func New(ctx context.Context) (context.Context, error) {
	envValue := ctx.Value(LoggerEnvKey)
	if envValue == nil {
		return nil, fmt.Errorf("missing logger environment variable")
	}
	env, ok := envValue.(string)
	if !ok {
		return nil, fmt.Errorf("invalid logger environment variable")
	}
	loggerStruct, err := NewLogger(env)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, LoggerKey, loggerStruct)
	return ctx, nil
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, loggerRequestIDKey, requestID)
}
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, loggerTraceIDKey, traceID)
}

func GetLogger(ctx context.Context) Logger {
	loggerValue := ctx.Value(LoggerKey)
	if loggerValue == nil {
		fallbackLogger, _ := NewLogger("info")
		return fallbackLogger
	}

	logger, ok := loggerValue.(Logger)
	if !ok {
		fallbackLogger, _ := NewLogger("info")
		return fallbackLogger
	}

	return logger
}

func TryAppendRequestID(ctx context.Context, fields []zap.Field) []zap.Field {
	if RequestID := ctx.Value(loggerRequestIDKey).(string); RequestID != "" {
		fields = append(fields, zap.String(loggerRequestIDKey, RequestID))
	}
	return fields
}

func TryAppendTraceID(ctx context.Context, fields []zap.Field) []zap.Field {
	if TraceID := ctx.Value(loggerTraceIDKey).(string); TraceID != "" {
		fields = append(fields, zap.String(loggerTraceIDKey, TraceID))
	}
	return fields
}

func (l *L) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = TryAppendRequestID(ctx, fields)
	fields = TryAppendTraceID(ctx, fields)

	l.z.Info(msg, fields...)
}
func (l *L) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = TryAppendRequestID(ctx, fields)
	fields = TryAppendTraceID(ctx, fields)

	l.z.Error(msg, fields...)
}
func (l *L) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = TryAppendRequestID(ctx, fields)
	fields = TryAppendTraceID(ctx, fields)

	l.z.Debug(msg, fields...)
}

func (l *L) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = TryAppendRequestID(ctx, fields)
	fields = TryAppendTraceID(ctx, fields)

	l.z.Fatal(msg, fields...)
}

func (l *L) Sync() error {
	return l.z.Sync()
}
