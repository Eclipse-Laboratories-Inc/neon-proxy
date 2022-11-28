package service

import (
	"fmt"

	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

type LoggerManager struct {
	loggers map[string]logger.Logger
}

func NewLoggerManager(log logger.Logger) *LoggerManager {
	loggerManager := &LoggerManager{
		loggers: make(map[string]logger.Logger),
	}
	loggerManager.setDefaultLogger(log)

	return loggerManager
}

func (l *LoggerManager) setDefaultLogger(logger logger.Logger) {
	l.loggers["default"] = logger
}

func (l *LoggerManager) SetLogger(name string, logger logger.Logger) {
	l.loggers[name] = logger
}

func (l *LoggerManager) GetLogger() logger.Logger {
	log, ok := l.loggers["default"]

	if !ok {
		panic("default logger not found")
	}

	return log
}

func (l *LoggerManager) GetLoggerByName(name string) logger.Logger {
	log, ok := l.loggers[name]

	if !ok {
		panic(fmt.Sprintf("logger with name %s not found", name))
	}

	return log
}
