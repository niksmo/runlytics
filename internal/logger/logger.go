package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log = zap.NewNop()

func Init(lvl string) {
	atomicLvl, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		panic(err)
	}

	config := zap.NewDevelopmentConfig()
	config.Level = atomicLvl
	config.Encoding = "json"
	config.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder

	if Log, err = config.Build(); err != nil {
		panic(err)
	}
}
