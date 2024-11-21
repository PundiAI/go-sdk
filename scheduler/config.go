package scheduler

import (
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	Enabled     bool          `yaml:"enabled" mapstructure:"enabled"`
	Name        string        `yaml:"name" mapstructure:"name"`
	Interval    time.Duration `yaml:"task_interval" mapstructure:"task_interval"`
	MaxErrCount uint16        `yaml:"task_max_err_count" mapstructure:"task_max_err_count"`
}

func NewDefConfig() Config {
	return Config{
		Enabled:     true,
		Interval:    time.Second * 5,
		MaxErrCount: 10,
	}
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Interval < 100*time.Millisecond || c.Interval > time.Second*600 {
		return errors.Errorf("task_interval is invalid, must between 100ms and 600s, got: %s", c.Interval.String())
	}
	if c.MaxErrCount <= 0 || c.MaxErrCount > 10000 {
		return errors.Errorf("task_max_err_count is invalid, must between 1 and 10000, got: %d", c.MaxErrCount)
	}
	return nil
}
