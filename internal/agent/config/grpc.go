package config

type GRPCConfig struct {
	IsSet bool
}

func NewGRPCConfig(p ConfigParams) (c GRPCConfig) {
	switch {
	case p.EnvSet.IsSet(grpcEnvName):
		c.IsSet = *p.EnvValues.grpc
	case p.FlagSet.IsSet(grpcFlagName):
		c.IsSet = *p.FlagValues.grpc
	case p.Settings.GRPC != nil:
		c.IsSet = *p.Settings.GRPC
	default:
		c.IsSet = grpcDefault
	}

	return
}
