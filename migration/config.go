package migration

import (
	"fmt"

	"github.com/pundiai/go-sdk/server"
)

var _ server.Config = Config{}

type Config struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

func NewDefConfig() Config {
	return Config{
		Enabled: false,
	}
}

func (c Config) String() string {
	return fmt.Sprintf("enabled: %t", c.Enabled)
}

func (c Config) IsEnabled() bool {
	return c.Enabled
}

func (c Config) Check() error {
	return nil
}

func (c Config) Name() string {
	return "migration"
}
