package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(level string) *zap.Logger {
	var l zapcore.Level
	if err := l.Set(level); err != nil {
		l = zapcore.InfoLevel
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(l)
	logger, _ := cfg.Build()
	return logger
}
