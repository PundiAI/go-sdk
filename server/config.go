package server

func NewDefConfig() BaseConfig {
	return BaseConfig{
		Enabled: true,
	}
}

type BaseConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}
