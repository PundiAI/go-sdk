package db

import (
	"context"
	"database/sql"

	mysql2 "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4/database"
	mysql3 "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/pundiai/go-sdk/log"
)

const MysqlDriver = "mysql"

func init() {
	RegisterDriver(MysqlDriver, &Mysql{})
}

var _ Driver = (*Mysql)(nil)

type Mysql struct {
	config *mysql2.Config
}

func (m *Mysql) ToMigrateDriver(source string) (string, database.Driver, error) {
	db, err := sql.Open(MysqlDriver, source)
	if err != nil {
		return "", nil, errors.Wrap(err, "mysql: open db error")
	}
	driver, err := mysql3.WithInstance(db, &mysql3.Config{})
	if err != nil {
		return "", nil, errors.Wrap(err, "mysql: with instance error")
	}
	return m.GetDatabaseName(source), driver, nil
}

func (m *Mysql) Open(source string) gorm.Dialector {
	return mysql.Open(source)
}

func (m *Mysql) ParseSource(source string) error {
	if m.config != nil {
		return nil
	}
	if source == "" {
		return errors.New("mysql: source is empty")
	}
	config, err := mysql2.ParseDSN(source)
	if err != nil {
		return errors.Wrap(err, "mysql: parse dsn error")
	}
	if config.DBName == "" {
		return errors.New("mysql: db name is empty")
	}
	if config.User == "" {
		return errors.New("mysql: db user is empty")
	}
	if config.Passwd == "" {
		return errors.New("mysql: db password is empty")
	}
	if config.Addr == "" {
		return errors.New("mysql: db address is empty")
	}
	if config.Net == "" {
		return errors.New("mysql: db net is empty")
	}
	m.config = config
	return nil
}

func (m *Mysql) GetDatabaseName(source string) string {
	if err := m.ParseSource(source); err != nil {
		panic(err)
	}
	return m.config.DBName
}

func (m *Mysql) getInformationDB(logger log.Logger, config Config) (DB, error) {
	if err := m.ParseSource(config.Source); err != nil {
		return nil, err
	}
	c := *m.config
	c.DBName = "information_schema"
	config.Source = c.FormatDSN()
	db, err := NewDB(context.Background(), logger, config)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (m *Mysql) CreateDB(logger log.Logger, config Config) error {
	db, err := m.getInformationDB(logger, config)
	if err != nil {
		return err
	}
	databaseName := config.GetDatabaseName()
	createDbSQL := "CREATE DATABASE IF NOT EXISTS " + databaseName + " DEFAULT CHARSET utf8 COLLATE utf8_general_ci;"
	if err = db.Exec(createDbSQL); err != nil {
		return err
	}
	logger.Info("create database success", "database", databaseName)
	return nil
}

func (m *Mysql) DropDB(logger log.Logger, config Config) error {
	db, err := m.getInformationDB(logger, config)
	if err != nil {
		return err
	}
	databaseName := config.GetDatabaseName()
	dropDbSQL := "DROP DATABASE IF EXISTS " + databaseName + ";"
	if err = db.Exec(dropDbSQL); err != nil {
		return err
	}
	logger.Info("drop database success", "database", databaseName)
	return nil
}

func (m *Mysql) GetMigrationsDriver() (source.Driver, error) {
	return GetMigrationsDriver(MysqlDriver)
}

func (m *Mysql) MigrateOptions() map[string]string {
	return map[string]string{
		"gorm:table_options": "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
	}
}
