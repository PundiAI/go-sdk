package db_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/pundiai/go-sdk/db"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestNewDefConfig() {
	config := db.NewDefConfig()
	config.Source = "root:root@tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local"
	suite.Equal(`driver: sqlite
source: '****:****@tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local'
conn_max_idle_time: 1h0m0s
conn_max_life_time: 1h0m0s
max_idle_conn: 10
max_open_conn: 30
log_level: silent
enable_metric: true
refresh_metric_interval: 15s
`, config.String())
}

func (suite *ConfigTestSuite) TestValidate() {
	config := db.NewDefConfig()
	suite.Require().NoError(config.Validate())

	config.Driver = ""
	suite.EqualError(config.Validate(), "driver is empty")
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func TestSourceDesensitization(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "test1",
			source: "my.db",
			want:   "my.db",
		},
		{
			name:   "test2",
			source: "root:root@tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local",
			want:   "****:****@tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name:   "test3",
			source: "root@tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local",
			want:   "*:*@tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name:   "test4",
			source: "tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local",
			want:   "tcp(127.0.0.1:3306)/my?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := db.SourceDesensitization(tt.source); got != tt.want {
				t.Errorf("SourceDesensitization() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonMarshalIndent(t *testing.T) {
	_, err := json.MarshalIndent(db.NewDefConfig(), "", "  ")
	require.NoError(t, err)
}
