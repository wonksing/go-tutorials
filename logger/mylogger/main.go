package main

import (
	"context"
	"log"
	"os"

	"github.com/wonksing/go-tutorials/logger/mylogger/logger"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/types"
)

func main() {
	roller, err := logger.RollerFactory(types.LumberjackRoller, "logs/agent.log", 1, 3, 3, true)
	if err != nil {
		log.Println("roller:", err)
		os.Exit(1)
	}
	_, err = logger.LoggerFactory(types.ZapLoggerType,
		logger.WithLevel(types.DebugLevel),
		logger.WithStdOut(true),
		logger.WithRoller(roller))
	if err != nil {
		log.Println("logger:", err)
		os.Exit(1)
	}
	logger.SetServiceName("myloggerApp")

	ctx := context.Background()
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message")
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")
	// logger.Fatal(ctx, "fatal message")
	// logger.Panic(ctx, "panic message")

	log.Println("finished")
	// log.SetOutput(myLogger)
}
