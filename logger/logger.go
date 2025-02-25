package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	instance *zap.Logger
	once     sync.Once
)

func GetLogger() *zap.Logger {
	once.Do(func() {
		logger, _ := zap.NewProduction()
		defer logger.Sync()
		instance = logger
	})

	return instance
}
