package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init(level string) {
	var config zap.Config

	if level == "debug" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic("日志初始化失败: " + err.Error())
	}

	Logger.Info("日志初始化成功", zap.String("level", level))
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
