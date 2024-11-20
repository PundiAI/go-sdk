package telemetry

import (
	"time"

	"github.com/pkg/errors"

	"github.com/pundiai/go-sdk/server"
	"github.com/pundiai/go-sdk/version"
)

var _ server.Config = Config{}

// Config defines the configuration options for application telemetry.
type Config struct {
	// Prefixed with keys to separate services
	ServiceName string `yaml:"service_name" mapstructure:"service_name"`

	// Enabled enables the application telemetry functionality. When enabled,
	// an in-memory sink is also enabled by default. Operators may also enabled
	// other sinks such as Prometheus.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// Address to listen for Prometheus collector(s) connections.
	ListenAddr string `yaml:"listen_addr" mapstructure:"listen_addr"`

	// ReadTimeout is the maximum duration for reading the entire request,
	ReadTimeout time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`

	// Maximum number of simultaneous connections.
	// If you want to accept a larger number than the default, make sure
	// you increase your OS limits.
	// 0 - unlimited.
	MaxOpenConnections int `yaml:"max_open_connections" mapstructure:"max_open_connections"`

	// GlobalLabels defines a global set of name/value label tuples applied to all
	// metrics emitted using the wrapper functions defined in telemetry package.
	GlobalLabels [][]string `yaml:"global_labels" mapstructure:"global_labels"`

	Interval time.Duration `yaml:"interval" mapstructure:"interval"`
}

func NewDefConfig() Config {
	return Config{
		ServiceName:        version.Name,
		Enabled:            true,
		ListenAddr:         "localhost:8080",
		ReadTimeout:        5 * time.Second,
		MaxOpenConnections: 3,
		GlobalLabels:       [][]string{},
		Interval:           30 * time.Second,
	}
}

func (c Config) IsEnabled() bool {
	return c.Enabled
}

func (c Config) Check() error {
	if !c.Enabled {
		return nil
	}
	if c.ListenAddr == "" {
		return errors.New("check: listen addr is empty")
	}
	if c.ReadTimeout < time.Millisecond {
		return errors.New("check: read timeout is too small")
	}
	if c.MaxOpenConnections <= 0 {
		return errors.New("check: max open connections is negative")
	}
	return nil
}

func (Config) Name() string {
	return "telemetry"
}
