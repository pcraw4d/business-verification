package main

import (
	"go.uber.org/zap"
)

// cacheLoggerWrapper wraps zap.Logger to implement cache.Logger interface
type cacheLoggerWrapper struct {
	logger *zap.Logger
}

func (c *cacheLoggerWrapper) Info(msg string, args ...interface{}) {
	c.logger.Info(msg, zap.Any("args", args))
}

func (c *cacheLoggerWrapper) Error(msg string, args ...interface{}) {
	c.logger.Error(msg, zap.Any("args", args))
}

func (c *cacheLoggerWrapper) Warn(msg string, args ...interface{}) {
	c.logger.Warn(msg, zap.Any("args", args))
}

func (c *cacheLoggerWrapper) Debug(msg string, args ...interface{}) {
	c.logger.Debug(msg, zap.Any("args", args))
}
