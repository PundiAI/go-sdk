package db_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/pundiai/go-sdk/db"
	"github.com/pundiai/go-sdk/log"
)

type DBTestSuite struct {
	suite.Suite
	db db.DB
}

func (suite *DBTestSuite) SetupTest() {
	suite.db = db.NewMemoryDB(log.LevelError, "db-test")
	suite.Require().NoError(suite.db.AutoMigrate(&gorm.Model{}))
}

func (suite *DBTestSuite) TestExec() {
	suite.Require().Error(suite.db.Exec("123"))
}

func (suite *DBTestSuite) TestFirst() {
	module := &gorm.Model{}
	found, err := suite.db.First(module)
	suite.Require().NoError(err)
	suite.False(found)
	suite.NotNil(module)
}

func (suite *DBTestSuite) TestFirstWhere() {
	found, err := suite.db.Where("id", 1).First(&gorm.Model{})
	suite.Require().NoError(err)
	suite.False(found)
}

func (suite *DBTestSuite) TestForWhere() {
	for i := 0; i < 10; i++ {
		found, err := suite.db.Where("id", 1).First(&gorm.Model{})
		suite.Require().NoError(err)
		suite.False(found)
	}
}

func (suite *DBTestSuite) TestFind() {
	modules := make([]gorm.Model, 0)
	err := suite.db.Find(&modules)
	suite.Require().NoError(err)
	suite.Empty(modules)
}

func (suite *DBTestSuite) TestDelete() {
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 1}))
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 2}))
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 3}))
	err := suite.db.Delete(&gorm.Model{}, "id", 1)
	suite.Require().NoError(err)
	var found bool
	found, err = suite.db.Where("id", 1).First(&gorm.Model{})
	suite.Require().NoError(err)
	suite.False(found)
	found, err = suite.db.Where("id", 2).First(&gorm.Model{})
	suite.Require().NoError(err)
	suite.True(found)
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

func (suite *DBTestSuite) TestTransaction() {
	testCases := []struct {
		name   string
		malloc func(tx db.DB) error
		err    error
	}{
		{
			name: "SUCCESS",
			malloc: func(tx db.DB) error {
				err := tx.Create(&gorm.Model{ID: 1})
				suite.Require().NoError(err)

				err = tx.Create(&gorm.Model{ID: 2})
				suite.Require().NoError(err)

				err = tx.Create(&gorm.Model{ID: 3})
				suite.Require().NoError(err)

				return nil
			},
			err: nil,
		},
		{
			name: "rollback",
			malloc: func(tx db.DB) error {
				err := tx.Create(&gorm.Model{ID: 4})
				suite.Require().NoError(err)

				err = tx.Create(&gorm.Model{ID: 5})
				suite.Require().NoError(err)

				err = tx.Create(&gorm.Model{ID: 6})
				suite.Require().NoError(err)
				return errors.New("rollback")
			},
			err: errors.New("rollback"),
		},
	}

	for _, tt := range testCases {
		suite.Run(tt.name, func() {
			malloc := tt.malloc
			err := suite.db.Transaction(malloc)
			var found bool
			if err == nil && tt.err == nil {
				suite.Require().NoError(err)
				found, err = suite.db.Where("id", 1).First(&gorm.Model{})
				suite.Require().NoError(err)
				suite.True(found)
				found, err = suite.db.Where("id", 2).First(&gorm.Model{})
				suite.Require().NoError(err)
				suite.True(found)
				found, err = suite.db.Where("id", 3).First(&gorm.Model{})
				suite.Require().NoError(err)
				suite.True(found)
			} else {
				suite.Require().Error(err)
				found, err = suite.db.Where("id", 4).First(&gorm.Model{})
				suite.Require().NoError(err)
				suite.False(found)
				found, err = suite.db.Where("id", 5).First(&gorm.Model{})
				suite.Require().NoError(err)
				suite.False(found)
				found, err = suite.db.Where("id", 6).First(&gorm.Model{})
				suite.Require().NoError(err)
				suite.False(found)
			}
		})
	}
}

func (suite *DBTestSuite) TestGDB_Scopes() {
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 1, CreatedAt: time.Now()}))
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 2, CreatedAt: time.Now()}))
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 3, CreatedAt: time.Now()}))
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 4, CreatedAt: time.Now()}))
	suite.Require().NoError(suite.db.Create(&gorm.Model{ID: 5, CreatedAt: time.Now()}))

	moreThenOne := func(db db.DB) db.DB {
		return db.Where("id > ?", 1)
	}
	lessThree := func(db db.DB) db.DB {
		return db.Where("id < ?", 3)
	}
	var model []*gorm.Model
	suite.Require().NoError(suite.db.Scopes(moreThenOne, lessThree).Find(&model))
	suite.T().Logf("%+v", model)

	more := func(id []uint) func(db db.DB) db.DB {
		return func(db db.DB) db.DB {
			return db.Where("id IN (?)", id)
		}
	}
	suite.Require().NoError(suite.db.Scopes(more([]uint{3, 4, 5})).Find(&model))
	suite.T().Logf("%+v", model)
}
