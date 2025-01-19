package logger

import (
	"go.uber.org/zap"
)

var Log = zap.NewNop()

func Initialize(lvl string) error {
	atomicLvl, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		return err
	}

	config := zap.NewDevelopmentConfig()
	config.Level = atomicLvl

	if Log, err = config.Build(); err != nil {
		return err
	}

	return nil
}
