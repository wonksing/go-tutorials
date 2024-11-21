package logger

import (
	"context"
	"fmt"

	"github.com/wonksing/go-tutorials/logger/mylogger/logger/port"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/types"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/wrapper"
)

var (
	_globalLogger port.Logger
	_serviceName  string
)

func SetServiceName(serviceName string) {
	_serviceName = serviceName
}

func Debug(ctx context.Context, message string, fields ...types.Field) {
	_globalLogger.Debug(ctx, message, fields...)
}

func Info(ctx context.Context, message string, fields ...types.Field) {
	_globalLogger.Info(ctx, message, fields...)
}

func Warn(ctx context.Context, message string, fields ...types.Field) {
	_globalLogger.Warn(ctx, message, fields...)
}

func Error(ctx context.Context, message string, fields ...types.Field) {
	_globalLogger.Error(ctx, message, fields...)
}

func Fatal(ctx context.Context, message string, fields ...types.Field) {
	_globalLogger.Fatal(ctx, message, fields...)
}

func Panic(ctx context.Context, message string, fields ...types.Field) {
	_globalLogger.Panic(ctx, message, fields...)
}

type Logger struct {
	l port.Logger

	roller port.Roller
	level  types.Level
	stdOut bool
}

func LoggerFactory(loggerType types.LoggerType, opts ...LoggerOption) (port.Logger, error) {
	switch loggerType {
	case types.ZapLoggerType:
		logger := &Logger{}
		for _, opt := range opts {
			opt.apply(logger)
		}

		zl, err := wrapper.NewZapLogger(logger.level, logger.stdOut, logger.roller)
		if err != nil {
			return nil, err
		}
		logger.l = zl

		_globalLogger = logger

		return logger, nil
	default:
		return nil, fmt.Errorf("unknown logger type: %d", loggerType)
	}
}

func (l *Logger) Debug(ctx context.Context, message string, fields ...types.Field) {
	fields = _appendFields(fields...)
	l.l.Debug(ctx, message, fields...)
}

func (l *Logger) Info(ctx context.Context, message string, fields ...types.Field) {
	fields = _appendFields(fields...)
	l.l.Info(ctx, message, fields...)
}

func (l *Logger) Warn(ctx context.Context, message string, fields ...types.Field) {
	fields = _appendFields(fields...)
	l.l.Warn(ctx, message, fields...)
}

func (l *Logger) Error(ctx context.Context, message string, fields ...types.Field) {
	fields = _appendFields(fields...)
	l.l.Error(ctx, message, fields...)
}

func (l *Logger) Fatal(ctx context.Context, message string, fields ...types.Field) {
	fields = _appendFields(fields...)
	l.l.Fatal(ctx, message, fields...)
}

func (l *Logger) Panic(ctx context.Context, message string, fields ...types.Field) {
	fields = _appendFields(fields...)
	l.l.Panic(ctx, message, fields...)
}

func _appendFields(fields ...types.Field) []types.Field {
	if _serviceName != "" {
		fields = append(fields, ServiceNameField(_serviceName))
	}
	return fields
}

func ServiceNameField(value string) types.Field {
	if value == "" {
		return types.Field{}
	}
	return types.WithStringField("service_name", value)
}
