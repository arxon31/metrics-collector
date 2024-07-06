package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func init() {
	Logger = initLogger()
}

func initLogger() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}
