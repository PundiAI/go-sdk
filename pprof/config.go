package pprof

import (
	"time"

	"github.com/pkg/errors"

	"github.com/pundiai/go-sdk/server"
)

var _ server.Config = Config{}

type Config struct {
	server.BaseConfig
	ListenAddr  string        `yaml:"listen_addr" mapstructure:"listen_addr"`
	ReadTimeout time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
}

func NewDefConfig() Config {
	return Config{
		BaseConfig:  server.NewDefConfig(),
		ListenAddr:  "localhost:6060",
		ReadTimeout: 5 * time.Second,
	}
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.ListenAddr == "" {
		return errors.New("listen addr is empty")
	}
	if c.ReadTimeout < time.Millisecond {
		return errors.New("read timeout is too small")
	}
	return nil
}

func (Config) Name() string {
	return "pprof"
}
