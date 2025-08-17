package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env, lvl string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "prod" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	if lvl != "" {
		if err := cfg.Level.UnmarshalText([]byte(lvl)); err != nil {
			return nil, err
		}
	}
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build()
}
