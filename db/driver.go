package db

import (
	"sync"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/pundiai/go-sdk/log"
)

var (
	drivers          = make(map[string]Driver)
	migrationDrivers = make(map[string]source.Driver)
	lock             = new(sync.RWMutex)
)

func RegisterDriver(name string, driver Driver) {
	lock.Lock()
	defer lock.Unlock()
	drivers[name] = driver
}

func GetDriver(name string) (Driver, error) {
	lock.RLock()
	defer lock.RUnlock()
	driver, ok := drivers[name]
	if !ok {
		return nil, errors.New("driver not support: " + name)
	}
	return driver, nil
}

type Driver interface {
	Open(source string) gorm.Dialector
	ParseSource(source string) error
	GetDatabaseName(source string) string
	CreateDB(logger log.Logger, config Config) error
	DropDB(logger log.Logger, config Config) error
	MigrateOptions() map[string]string
	GetMigrationsDriver() (source.Driver, error)
	ToMigrateDriver(source string) (string, database.Driver, error)
}

func RegisterMigrationsDriver(name string, driver source.Driver) {
	lock.Lock()
	defer lock.Unlock()
	migrationDrivers[name] = driver
}

func GetMigrationsDriver(name string) (source.Driver, error) {
	lock.RLock()
	defer lock.RUnlock()
	driver, ok := migrationDrivers[name]
	if !ok {
		return nil, errors.New("migration driver not support: " + name)
	}
	return driver, nil
}
