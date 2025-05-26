package config

type DBConfig struct {
	DSN string
}

func NewDBConfig(p ConfigParams) (c DBConfig) {
	switch {
	case p.EnvSet.IsSet(dsnEnvName):
		c.DSN = *p.EnvValues.dsn
	case p.FlagSet.IsSet(dsnFlagName):
		c.DSN = *p.FlagValues.dsn
	case p.Settings.DSN != nil:
		c.DSN = *p.Settings.DSN
	}
	return
}

func (c *DBConfig) IsSet() bool {
	return c.DSN != ""
}
