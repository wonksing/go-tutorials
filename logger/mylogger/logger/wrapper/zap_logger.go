package wrapper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/wonksing/go-tutorials/logger/mylogger/logger/port"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ port.Closer = (*ZapLogger)(nil)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(level types.Level, stdOut bool, roller port.Roller) (*ZapLogger, error) {

	var zlevel zapcore.Level = zapLevel(level)
	var err error

	config := zap.NewProductionConfig()
	config.Sampling = nil
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	// config.EncoderConfig.FunctionKey = "funcName"
	config.EncoderConfig.StacktraceKey = ""

	config.Level = zap.NewAtomicLevelAt(zlevel)
	config.Encoding = "json"
	config.OutputPaths = []string{}

	if roller != nil {
		zap.RegisterSink("roller", func(*url.URL) (zap.Sink, error) {
			return roller, nil
		})
		config.OutputPaths = append(config.OutputPaths, fmt.Sprintf("roller:%s", roller.GetPath()))
	}

	if stdOut {
		config.OutputPaths = append(config.OutputPaths, "stdout")
	}

	_logger, err := config.Build(zap.AddCallerSkip(2))
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(_logger)

	return &ZapLogger{
		logger: _logger,
	}, nil
}

func (z *ZapLogger) Close() error {
	if z.logger != nil {
		return z.logger.Sync()
	}
	return nil
}

func (z *ZapLogger) Debug(ctx context.Context, message string, fields ...types.Field) {
	z.logger.Debug(message, convertFields(fields...)...)
}

func (z *ZapLogger) Info(ctx context.Context, message string, fields ...types.Field) {
	z.logger.Info(message, convertFields(fields...)...)
}

func (z *ZapLogger) Warn(ctx context.Context, message string, fields ...types.Field) {
	z.logger.Warn(message, convertFields(fields...)...)
}

func (z *ZapLogger) Error(ctx context.Context, message string, fields ...types.Field) {
	z.logger.Error(message, convertFields(fields...)...)
}

func (z *ZapLogger) Fatal(ctx context.Context, message string, fields ...types.Field) {
	z.logger.Fatal(message, convertFields(fields...)...)
}

func (z *ZapLogger) Panic(ctx context.Context, message string, fields ...types.Field) {
	z.logger.Panic(message, convertFields(fields...)...)
}

func convertFields(fields ...types.Field) []zapcore.Field {
	fieldLength := len(fields)
	if fieldLength == 0 {
		var empty []zapcore.Field
		return empty
	}

	var fieldIndex uint8 = 0
	zapFields := make([]zapcore.Field, fieldLength)

	for _, f := range fields {
		if f.Name == "" {
			continue
		}
		switch f.Type {
		case types.StringType:
			zapFields[fieldIndex] = zap.String(f.Name, f.ValueString)
			fieldIndex++
		case types.BytesType:
			zapFields[fieldIndex] = zap.ByteString(f.Name, f.ValueBytes)
			fieldIndex++
		case types.Int32Type:
			zapFields[fieldIndex] = zap.Int32(f.Name, f.ValueInt32)
			fieldIndex++
		case types.Int64Type:
			zapFields[fieldIndex] = zap.Int64(f.Name, f.ValueInt64)
			fieldIndex++
		case types.Uint32Type:
			zapFields[fieldIndex] = zap.Uint32(f.Name, f.ValueUint32)
			fieldIndex++
		case types.Uint64Type:
			zapFields[fieldIndex] = zap.Uint64(f.Name, f.ValueUint64)
			fieldIndex++
		case types.AnyType:
			zapFields[fieldIndex] = zap.Any(f.Name, f.ValueAny)
			fieldIndex++
		default:
			panic(fmt.Errorf("unknown field type, %d(%s)", f.Type, f.Name))
		}
	}

	return zapFields[:fieldIndex]
}

func zapLevel(level types.Level) zapcore.Level {
	var zlevel zapcore.Level
	switch level {
	case types.DebugLevel:
		zlevel = zap.DebugLevel
	case types.InfoLevel:
		zlevel = zap.InfoLevel
	case types.WarnLevel:
		zlevel = zap.WarnLevel
	case types.ErrorLevel:
		zlevel = zap.ErrorLevel
	case types.FatalLevel:
		zlevel = zap.FatalLevel
	case types.PanicLevel:
		zlevel = zap.PanicLevel
	default:
		panic(fmt.Errorf("unknown log level %s", level))
	}
	return zlevel
}
