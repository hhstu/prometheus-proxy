package log

import (
	"github.com/hhstu/prometheus-proxy/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func init() {
	logLevel := zap.DebugLevel
	switch config.AppConfig.Log.Level {
	case "debug", "DEBUG":
		logLevel = zap.DebugLevel
	case "info", "INFO":
		logLevel = zap.InfoLevel
	case "warn", "WARN":
		logLevel = zap.WarnLevel
	case "error", "ERROR":
		logLevel = zap.ErrorLevel
	}
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "line",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	if config.AppConfig.Log.Encoding != "" {
		cfg.Encoding = config.AppConfig.Log.Encoding
		if config.AppConfig.Log.Encoding == "console" {
			cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		}
	}
	Logger = zap.Must(cfg.Build()).Sugar()
}
