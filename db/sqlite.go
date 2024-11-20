package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4/database"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pundiai/go-sdk/log"
)

const SqliteDriver = "sqlite"

func init() {
	RegisterDriver(SqliteDriver, &Sqlite{})
}

var _ Driver = (*Sqlite)(nil)

type Sqlite struct {
	name string
}

func (*Sqlite) Open(source string) gorm.Dialector {
	if err := os.MkdirAll(filepath.Dir(source), os.ModePerm); err != nil {
		panic(err)
	}
	return sqlite.Open(source)
}

func (s *Sqlite) ParseSource(source string) error {
	if source == "" {
		return errors.New("sqlite: db name is empty")
	}
	match := regexp.MustCompile(`file:([^\.]+)\.db`).FindStringSubmatch(source)
	if len(match) > 1 {
		s.name = match[1]
		return nil
	}

	if !strings.HasSuffix(source, ".db") {
		return errors.New("sqlite: db name suffix must be .db")
	}
	s.name = strings.TrimSuffix(filepath.Base(source), ".db")
	return nil
}

func (s *Sqlite) GetDatabaseName(source string) string {
	if err := s.ParseSource(source); err != nil {
		panic(err)
	}
	return s.name
}

func (*Sqlite) CreateDB(logger log.Logger, config Config) error {
	if err := os.MkdirAll(config.Source, os.ModePerm); err != nil {
		return errors.Wrap(err, "sqlite: create db error")
	}
	logger.Info("sqlite: create db success", "source", config.Source)
	return nil
}

func (*Sqlite) DropDB(logger log.Logger, config Config) error {
	if err := os.RemoveAll(config.Source); err != nil {
		return errors.Wrap(err, "sqlite: drop db error")
	}
	logger.Info("sqlite: drop db success", "source", config.Source)
	return nil
}

func (*Sqlite) MigrateOptions() map[string]string {
	return map[string]string{}
}

func (*Sqlite) GetMigrationsDriver() (source.Driver, error) {
	return GetMigrationsDriver(SqliteDriver)
}

func (*Sqlite) ToMigrateDriver(source string) (string, database.Driver, error) {
	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return "", nil, errors.Wrap(err, "sqlite: open error")
	}
	driver, err := migratesqlite.WithInstance(db, &migratesqlite.Config{})
	if err != nil {
		return "", nil, errors.Wrap(err, "sqlite: with instance error")
	}
	return "", driver, nil
}
