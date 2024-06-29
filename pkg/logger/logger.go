package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init() {
	Logger = initLogger()
}

func initLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}
