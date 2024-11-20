package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm/logger"
)

const (
	LogLevelSilent = "silent"
	LogLevelError  = "error"
	LogLevelWarn   = "warn"
	LogLevelInfo   = "info"
)

type Config struct {
	Driver          string        `yaml:"driver" mapstructure:"driver"`
	Source          string        `yaml:"source" mapstructure:"source"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time" mapstructure:"conn_max_life_time"`
	MaxIdleConn     int           `yaml:"max_idle_conn" mapstructure:"max_idle_conn"`
	MaxOpenConn     int           `yaml:"max_open_conn" mapstructure:"max_open_conn"`
	LogLevel        string        `yaml:"log_level" mapstructure:"log_level"`

	EnableMetric          bool          `yaml:"enable_metric" mapstructure:"enable_metric"`
	RefreshMetricInterval time.Duration `yaml:"refresh_metric_interval" mapstructure:"refresh_metric_interval"`
}

func NewDefConfig() Config {
	return Config{
		Driver:          SqliteDriver,
		Source:          os.ExpandEnv("$HOME/.my/my.db"),
		ConnMaxIdleTime: time.Hour,
		ConnMaxLifeTime: time.Hour,
		MaxIdleConn:     10,
		MaxOpenConn:     30,
		LogLevel:        LogLevelSilent,

		EnableMetric:          true,
		RefreshMetricInterval: 15 * time.Second,
	}
}

func (c Config) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (c Config) MarshalYAML() (any, error) {
	type marshalConfig Config
	temp := marshalConfig(c)
	temp.Source = SourceDesensitization(temp.Source)
	return temp, nil
}

func (c Config) MarshalJSON() ([]byte, error) {
	temp := c
	temp.Source = SourceDesensitization(temp.Source)
	return json.Marshal(temp)
}

func (c Config) Check() error { //nolint:revive // cyclomatic
	if c.Driver == "" {
		return errors.New("check: driver is empty")
	}
	driver, err := GetDriver(c.Driver)
	if err != nil {
		return errors.WithMessage(err, "check: driver is invalid")
	}
	if err = driver.ParseSource(c.Source); err != nil {
		return errors.WithMessage(err, "check: source is invalid")
	}
	if c.ConnMaxIdleTime < time.Second || c.ConnMaxIdleTime > time.Hour*24 {
		return errors.New("check: conn_max_idle_time is invalid, must between 1 seconds and 24 hours")
	}
	if c.ConnMaxLifeTime < time.Second || c.ConnMaxLifeTime > time.Hour*24 {
		return errors.New("check: conn_max_life_time is invalid, must between 1 seconds and 24 hours")
	}
	if c.MaxIdleConn < 1 || c.MaxIdleConn > 500 {
		return errors.New("check: max_idle_conn is invalid, must between 1 and 500")
	}
	if c.MaxOpenConn < 1 || c.MaxOpenConn > 500 {
		return errors.New("check: max_open_conn is invalid, must between 1 and 500")
	}
	if c.MaxOpenConn < c.MaxIdleConn {
		return errors.New("check: max_open_conn must greater than max_idle_conn")
	}
	if _, err = parseLogLevel(c.LogLevel); err != nil {
		return errors.WithMessage(err, "check: log_level is invalid")
	}
	return nil
}

func (c Config) GetLogLevel() logger.LogLevel {
	logLevel, _ := parseLogLevel(c.LogLevel)
	return logLevel
}

func (c Config) GetDatabaseName() string {
	driver, err := GetDriver(c.Driver)
	if err != nil {
		panic(err)
	}
	return driver.GetDatabaseName(c.Source)
}

func SourceDesensitization(source string) string {
	dbSourceArr := strings.Split(source, "@")
	if len(dbSourceArr) <= 1 {
		return source
	}
	showUserAndPassword := "*:*"
	userAndPassword := strings.Split(dbSourceArr[0], ":")
	if len(userAndPassword) == 2 {
		showUserAndPassword = fmt.Sprintf(
			"%s:%s",
			strings.Repeat("*", len(userAndPassword[0])),
			strings.Repeat("*", len(userAndPassword[1])),
		)
	}
	return fmt.Sprintf("%s@%s", showUserAndPassword, dbSourceArr[1])
}

func parseLogLevel(logLevel string) (logger.LogLevel, error) {
	switch logLevel {
	case LogLevelSilent:
		return logger.Silent, nil
	case LogLevelError:
		return logger.Error, nil
	case LogLevelWarn:
		return logger.Warn, nil
	case LogLevelInfo:
		return logger.Info, nil
	default:
		return logger.Silent, errors.New("invalid log level")
	}
}
