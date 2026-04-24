package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(mode string) {
	var cfg zap.Config
	if mode == "prod" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	log, _ = cfg.Build()
}

func Info(msg string, fields ...zap.Field)  { log.Info(msg, fields...) }
func Error(msg string, fields ...zap.Field) { log.Error(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { log.Warn(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { log.Fatal(msg, fields...) }
func Sync()                                 { _ = log.Sync() }
