package config

type HashKeyConfig struct {
	Key string
}

func NewHashKeyConfig(p ConfigParams) (hc HashKeyConfig) {
	switch {
	case p.EnvSet.IsSet(hashKeyEnvName):
		hc.Key = *p.EnvValues.hashKey
	case p.FlagSet.IsSet(hashKeyFlagName):
		hc.Key = *p.FlagValues.hashKey
	case p.Settings.HashKey != nil:
		hc.Key = *p.Settings.HashKey
	}
	return
}

func (hc HashKeyConfig) IsSet() bool {
	return hc.Key != ""
}
