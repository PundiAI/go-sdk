package db_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pundiai/go-sdk/db"
	"github.com/pundiai/go-sdk/log"
)

func TestSqlite_CheckSource(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "test1",
			source:  "my.db",
			wantErr: assert.NoError,
		},
		{
			name:    "test2",
			source:  "",
			wantErr: assert.Error,
		},
		{
			name:    "test3",
			source:  "my",
			wantErr: assert.Error,
		},
		{
			name:    "test4",
			source:  "file:my.db?mode=memory",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &db.Sqlite{}
			tt.wantErr(t, s.ParseSource(tt.source), fmt.Sprintf("ParseSource(%v)", tt.source))
		})
	}
}

type SqliteTestSuite struct {
	suite.Suite
	driver *db.Sqlite
}

func (suite *SqliteTestSuite) SetupTest() {
	suite.driver = new(db.Sqlite)
}

func (suite *SqliteTestSuite) TestOpen() {
	suite.NotNil(suite.driver.Open("my.db"))
}

func (suite *SqliteTestSuite) TestOpen2() {
	source := "${HOME}/.my/my.db"
	suite.T().Log(os.ExpandEnv(source))
	suite.NotNil(suite.driver.Open(os.ExpandEnv(source)))
}

func (suite *SqliteTestSuite) TestGetDatabaseName() {
	source := "my.db"
	suite.Equal("my", suite.driver.GetDatabaseName(source))

	source = suite.T().TempDir() + "/my.db"
	suite.Equal("my", suite.driver.GetDatabaseName(source))
}

func (suite *SqliteTestSuite) TestCreateDB() {
	source := suite.T().TempDir() + "/my.db"
	suite.Require().NoError(suite.driver.CreateDB(log.NewNopLogger(), db.Config{Source: source}))
	defer func() {
		suite.Require().NoError(suite.driver.DropDB(log.NewNopLogger(), db.Config{Source: source}))
	}()
	stat, err := os.Stat(source)
	suite.Require().NoError(err)
	suite.Require().NotNil(stat)
	suite.True(stat.IsDir())
	suite.Equal("my.db", stat.Name())
}

func TestSqliteTestSuite(t *testing.T) {
	suite.Run(t, new(SqliteTestSuite))
}
