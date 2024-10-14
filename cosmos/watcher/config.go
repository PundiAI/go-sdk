package watcher

import (
	"time"

	"github.com/pkg/errors"

	"github.com/pundiai/go-sdk/server"
)

var _ server.Config = Config{}

type Config struct {
	server.BaseConfig

	GrpcUrl string

	// BlockInterval is the interval between two blocks
	BlockInterval    time.Duration `yaml:"block_interval" mapstructure:"block_interval"`
	StartBlockHeight int64         `yaml:"start_block_height" mapstructure:"start_block_height"`
	EndBlockHeight   int64         `yaml:"end_block_height" mapstructure:"end_block_height"`
	BatchHandler     int           `yaml:"batch_handler" mapstructure:"batch_handler"`
}

func NewDefConfig() Config {
	return Config{
		BaseConfig:       server.NewDefConfig(),
		BlockInterval:    5 * time.Second,
		StartBlockHeight: -1,
		EndBlockHeight:   -1,
		BatchHandler:     100,
	}
}

func (c Config) Check() error {
	if !c.Enabled {
		return nil
	}
	if c.GrpcUrl == "" {
		return errors.New("check: grpc_url is empty")
	}
	if c.BatchHandler <= 0 {
		return errors.New("check: batch handler is invalid")
	}
	return nil
}

func (c Config) Name() string {
	return "watcher"
}
