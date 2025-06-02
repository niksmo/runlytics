package config

import (
	"fmt"
	"os"
	"time"
)

type FileStorageConfig struct {
	File         *os.File
	SaveInterval time.Duration
	Restore      bool
}

func NewFileStorageConfig(p ConfigParams) (fc FileStorageConfig) {
	fc.initFile(p)
	fc.initSaveInterval(p)
	fc.initRestore(p)
	return
}

func (fc *FileStorageConfig) FileName() string {
	if fc.File != nil {
		return fc.File.Name()
	}
	return ""
}

func (fc *FileStorageConfig) initFile(p ConfigParams) {
	resolveFile := func(path, src, name string) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			p.ErrStream <- fmt.Errorf(
				"failed to open store path '%s', source '%s' name '%s': %w",
				path, src, name, err,
			)
			return
		}
		fc.File = f
	}

	switch {
	case p.EnvSet.IsSet(storeEnvName):
		resolveFile(*p.EnvValues.store, srcEnv, storeEnvName)

	case p.FlagSet.IsSet(storeFlagName):
		resolveFile(*p.FlagValues.store, srcFlag, "-"+storeFlagName)

	case p.Settings.Restore != nil:
		resolveFile(*p.Settings.StoreFile, srcSettings, storeSettingsName)
	default:
		resolveFile(storeDefaultPath, "", "")
	}

}

func (fc *FileStorageConfig) initSaveInterval(p ConfigParams) {
	resolveSaveInterval := func(value int, src, name string) {
		if value < 0 {
			p.ErrStream <- fmt.Errorf(
				"store interval '%d' less zero, source '%s' name '%s'",
				value, src, name,
			)
			return
		}
		fc.SaveInterval = time.Second * time.Duration(value)
	}

	switch {
	case p.EnvSet.IsSet(storeIntervalEnvName):
		resolveSaveInterval(
			*p.EnvValues.storeInterval, srcEnv, storeIntervalEnvName,
		)
	case p.FlagSet.IsSet(storeIntervalFlagName):
		resolveSaveInterval(
			*p.FlagValues.storeInterval, srcFlag, "-"+storeIntervalFlagName,
		)
	case p.Settings.StoreInterval != nil:
		resolveSaveInterval(
			*p.Settings.StoreInterval, srcSettings, storeIntervalSettingsName,
		)
	default:
		resolveSaveInterval(storeIntervalDefault, "", "")
	}
}

func (fc *FileStorageConfig) initRestore(p ConfigParams) {
	switch {
	case p.EnvSet.IsSet(storeRestoreEnvName):
		fc.Restore = *p.EnvValues.storeRestore
	case p.FlagSet.IsSet(storeRestoreFlagName):
		fc.Restore = *p.FlagValues.storeRestore
	case p.Settings.Restore != nil:
		fc.Restore = *p.Settings.Restore
	}
}
