package dao_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/pundiai/go-sdk/dao"
	"github.com/pundiai/go-sdk/db"
	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/modules"
)

type TestModel struct {
	modules.Base `gorm:"embedded"`

	Name   string `gorm:"index:,unique; column:name; type:varchar(20); not null; comment:task name"`
	Number uint64 `gorm:"column:number; type:bigint(20);not null;comment:block number"`
}

func (v *TestModel) TableName() string {
	return "test_model"
}

func NewTestModel(name string, number uint64) *TestModel {
	return &TestModel{
		Name:   name,
		Number: number,
	}
}

type DaoTestSuite struct {
	suite.Suite
	baseDao *dao.BaseDao
}

func (s *DaoTestSuite) SetupTest() {
	s.doSetup()
}

func (s *DaoTestSuite) SetupSubTest() {
	s.doSetup()
}

func (s *DaoTestSuite) doSetup() {
	testDB := db.NewMemoryDB(log.LevelFatal, "base-dao-test")
	s.Require().NoError(testDB.AutoMigrate(
		new(TestModel),
	))
	s.baseDao = dao.NewDao(testDB, &TestModel{})
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, new(DaoTestSuite))
}

func (s *DaoTestSuite) TestInsert() {
	data := NewTestModel("test", 100)
	s.Require().NoError(s.baseDao.Insert(data))
	s.Require().NotZero(data.GetId(), "id should be set")
}

func (s *DaoTestSuite) TestGet() {
	data := NewTestModel("test", 100)
	s.Require().NoError(s.baseDao.Insert(data))
	s.Require().NotZero(data.GetId(), "id should be set")

	actualData := &TestModel{}
	found, err := s.baseDao.GetByID(data.GetId(), actualData)
	s.Require().NoError(err)
	s.Require().True(found)
	s.Require().Equal(data.GetId(), actualData.GetId())
	s.Require().Equal(data.Name, actualData.Name)
	s.Require().Equal(data.Number, actualData.Number)
}

func (s *DaoTestSuite) TestUpdatesByID() {
	data := NewTestModel("test", 100)
	s.Require().NoError(s.baseDao.Insert(data))
	s.Require().NotZero(data.GetId(), "id should be set")

	updateData := NewTestModel("test2", 200)
	s.Require().NoError(s.baseDao.UpdatesByID(data.GetId(), updateData))
	actualData := &TestModel{}
	found, err := s.baseDao.GetByID(data.GetId(), actualData)
	s.Require().NoError(err)
	s.Require().True(found)

	s.Require().Equal(data.GetId(), actualData.GetId())
	actualData.Base = modules.Base{}
	updateData.Base = modules.Base{}
	s.Require().EqualValues(updateData, actualData)
}

func (s *DaoTestSuite) TestNoTransaction() {
	func() {
		data := NewTestModel("test", 100)
		var txErr error
		defer func() {
			s.Require().NotNil(txErr)
			s.Require().EqualError(txErr, "db create error: UNIQUE constraint failed: test_model.name")
		}()
		if txErr = s.baseDao.Insert(data); txErr != nil {
			return
		}

		if _, txErr = s.baseDao.GetByID(data.GetId(), data); txErr != nil {
			return
		}
		// test insert duplicate
		if txErr = s.baseDao.Insert(NewTestModel("test", 100)); txErr != nil {
			return
		}
	}()
	var testModels []TestModel
	err := s.baseDao.GetDB().Find(&testModels)
	s.Require().NoError(err)
	s.Require().Len(testModels, 1)
}

func (s *DaoTestSuite) TestTransaction() {
	testCase := []struct {
		name        string
		insertData  []*TestModel
		rollbackErr string
		expectLen   int
	}{
		{
			name: "test commit",
			insertData: []*TestModel{
				NewTestModel("test1", 100),
				NewTestModel("test2", 200),
				NewTestModel("test3", 300),
			},
			rollbackErr: "",
			expectLen:   3,
		},
		{
			name: "test rollback - UNIQUE constraint",
			insertData: []*TestModel{
				NewTestModel("test1", 100),
				NewTestModel("test2", 200),
				NewTestModel("test1", 200),
			},
			rollbackErr: "db create error: UNIQUE constraint failed: test_model.name",
			expectLen:   0,
		},
	}
	for _, tc := range testCase {
		s.Run(tc.name, func() {
			func(data []*TestModel, rollbackErr string) {
				txDB := s.baseDao.GetDB().Begin()
				var txErr error
				defer func() {
					if txErr != nil {
						s.Require().EqualError(txErr, rollbackErr)
						s.Require().NoError(txDB.Rollback())
					}
				}()

				for _, d := range data {
					if txErr = txDB.Create(d); txErr != nil {
						return
					}
				}
				txErr = txDB.Commit()
			}(tc.insertData, tc.rollbackErr)

			var testModels []TestModel
			err := s.baseDao.GetDB().Find(&testModels)
			s.Require().NoError(err)
			s.Require().Len(testModels, tc.expectLen)
		})
	}
}

func (s *DaoTestSuite) TestTransactionWithCtx() {
	testCase := []struct {
		name       string
		insertData []*TestModel
		rollback   bool
		expectLen  int
	}{
		{
			name: "test commit",
			insertData: []*TestModel{
				NewTestModel("test1", 100),
				NewTestModel("test2", 200),
			},
			rollback:  false,
			expectLen: 2,
		},
		{
			name: "test rollback",
			insertData: []*TestModel{
				NewTestModel("test1", 100),
				NewTestModel("test2", 200),
			},
			rollback:  true,
			expectLen: 0,
		},
	}
	for _, tc := range testCase {
		s.Run(tc.name, func() {
			txCtx := s.baseDao.BeginTx(context.Background())
			for _, data := range tc.insertData {
				s.Require().NoError(s.baseDao.InsertWithCtx(txCtx, data))
			}
			if tc.rollback {
				s.Require().NoError(s.baseDao.RollbackTx(txCtx))
			} else {
				s.Require().NoError(s.baseDao.CommitTx(txCtx))
			}

			var count int64
			_ = s.baseDao.GetDB().Model(&TestModel{}).Count(&count)
			s.Require().Equal(int64(tc.expectLen), count)
		})
	}
}

func (s *DaoTestSuite) TestTransaction2WithCtx() {
	testCase := []struct {
		name       string
		insertData []*TestModel
		rollback   bool
		expectLen  int
	}{
		{
			name: "test commit",
			insertData: []*TestModel{
				NewTestModel("test1", 100),
				NewTestModel("test2", 200),
			},
			rollback:  false,
			expectLen: 2,
		},
		{
			name: "test rollback",
			insertData: []*TestModel{
				NewTestModel("test1", 100),
				NewTestModel("test2", 200),
			},
			rollback:  true,
			expectLen: 0,
		},
	}
	for _, tc := range testCase {
		s.Run(tc.name, func() {
			if tc.rollback {
				s.Require().Error(s.baseDao.Transaction(context.Background(), func(ctx context.Context) error {
					for _, data := range tc.insertData {
						s.Require().NoError(s.baseDao.InsertWithCtx(ctx, data))
					}
					return fmt.Errorf("rollback")
				}))
			} else {
				s.Require().NoError(s.baseDao.Transaction(context.Background(), func(ctx context.Context) error {
					for _, data := range tc.insertData {
						s.Require().NoError(s.baseDao.InsertWithCtx(ctx, data))
					}
					return nil
				}))
			}
			var count int64
			_ = s.baseDao.GetDB().Model(&TestModel{}).Count(&count)
			s.Require().Equal(int64(tc.expectLen), count)
		})
	}
}

func (s *DaoTestSuite) TestCallCommitMultipleTimes() {
	txCtx := s.baseDao.BeginTx(context.Background())
	s.Require().NoError(s.baseDao.CommitTx(txCtx))
	err := s.baseDao.CommitTx(txCtx)
	s.Require().Error(err)
	s.Require().EqualError(err, "sql: transaction has already been committed or rolled back")
}

func (s *DaoTestSuite) TestCallRollbackMultipleTimes() {
	txCtx := s.baseDao.BeginTx(context.Background())
	s.Require().NoError(s.baseDao.RollbackTx(txCtx))
	err := s.baseDao.RollbackTx(txCtx)
	s.Require().Error(err)
	s.Require().EqualError(err, "sql: transaction has already been committed or rolled back")
}

func (s *DaoTestSuite) TestCallRollbackAfterCommit() {
	txCtx := s.baseDao.BeginTx(context.Background())
	err := s.baseDao.CommitTx(txCtx)
	s.Require().NoError(err)
	err = s.baseDao.RollbackTx(txCtx)
	s.Require().Error(err)
	s.Require().EqualError(err, "sql: transaction has already been committed or rolled back")
}

func (s *DaoTestSuite) TestMultipleBeginTxAndCommit() {
	txCtx := s.baseDao.BeginTx(context.Background())

	txCtx2 := s.baseDao.BeginTx(txCtx)
	// expect txCtx2 commit or rollback will not affect txCtx
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))

	// only txCtx can commit
	s.Require().NoError(s.baseDao.CommitTx(txCtx))
}

func (s *DaoTestSuite) TestMultipleBeginTxAndRollback() {
	txCtx := s.baseDao.BeginTx(context.Background())

	txCtx2 := s.baseDao.BeginTx(txCtx)
	// expect txCtx2 commit or rollback will not affect txCtx
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.CommitTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))
	s.Require().NoError(s.baseDao.RollbackTx(txCtx2))

	// only txCtx can roll back
	s.Require().NoError(s.baseDao.RollbackTx(txCtx))
}

func (s *DaoTestSuite) TestMultipleTransactionTx() {
	s.Require().NoError(s.baseDao.Transaction(context.Background(), func(ctx1 context.Context) error {
		// commit
		s.Require().NoError(s.baseDao.Transaction(ctx1, func(ctx context.Context) error {
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test1", 100)))
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test2", 200)))
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test3", 300)))
			return nil
		}))
		// rollback
		s.Require().Error(s.baseDao.Transaction(ctx1, func(ctx context.Context) error {
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test4", 400)))
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test5", 500)))
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test6", 600)))
			return fmt.Errorf("rollback")
		}))
		// commit
		s.Require().NoError(s.baseDao.Transaction(ctx1, func(ctx context.Context) error {
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test7", 700)))
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test8", 800)))
			s.Require().NoError(s.baseDao.InsertWithCtx(ctx, NewTestModel("test9", 900)))
			return nil
		}))
		return s.baseDao.InsertWithCtx(ctx1, NewTestModel("test10", 1000))
	}))
	var count int64
	_ = s.baseDao.GetDB().Model(&TestModel{}).Count(&count)
	s.Require().Equal(int64(7), count)
}
