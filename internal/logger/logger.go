package logger

import (
	"go.uber.org/zap"
)

var Log = zap.NewNop()

func Init(lvl string) {
	atomicLvl, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		panic(err)
	}

	config := zap.NewDevelopmentConfig()
	config.Level = atomicLvl

	if Log, err = config.Build(); err != nil {
		panic(err)
	}
}
