package port

import (
	"context"

	"github.com/wonksing/go-tutorials/logger/mylogger/logger/types"
)

type Logger interface {
	Debug(ctx context.Context, message string, fields ...types.Field)
	Info(ctx context.Context, message string, fields ...types.Field)
	Warn(ctx context.Context, message string, fields ...types.Field)
	Error(ctx context.Context, message string, fields ...types.Field)
	Fatal(ctx context.Context, message string, fields ...types.Field)
	Panic(ctx context.Context, message string, fields ...types.Field)
}

func LoggerFactory(loggerType types.LoggerType, config string) (Logger, error) {
	return nil, nil
}
