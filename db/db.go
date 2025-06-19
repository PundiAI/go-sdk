package db

import (
	"context"
	"fmt"
	golog "log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
	"gorm.io/plugin/prometheus"

	"github.com/pundiai/go-sdk/log"
)

type DB interface {
	Model(value any) DB
	Where(query any, args ...any) DB
	Limit(limit int) DB
	Scopes(funcs ...func(DB) DB) DB
	Offset(offset int) DB
	Order(value any) DB
	Count(count *int64) DB
	Group(query string) DB
	RowsAffected(number int64) DB

	Select(query any, args ...any) DB
	Distinct(args ...any) DB
	Find(dest any, conds ...any) (err error)
	First(dest any, conds ...any) (found bool, err error)
	MustFirst(dest any, conds ...any) (err error)

	Exec(sql string, values ...any) error

	Create(value any) error
	Update(column string, value any) error
	Updates(values any) error
	Delete(value any, conds ...any) error

	Transaction(fn func(tx DB) error) error

	Begin() DB
	Commit() error
	Rollback() error

	AutoMigrate(dst ...any) error

	GetSource() string
	GetDriver() Driver
	Close() error

	WithContext(ctx context.Context) DB
	WithLogger(l log.Logger) DB
}

func NewDB(_ context.Context, l log.Logger, config Config) (DB, error) {
	driver, err := GetDriver(config.Driver)
	if err != nil {
		return nil, errors.WithMessage(err, "get driver error")
	}

	dblog := logger.New(
		golog.New(os.Stdout, "\r\n", golog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second * 2,      // Slow SQL threshold
			LogLevel:                  config.GetLogLevel(), // Log level
			IgnoreRecordNotFoundError: true,                 // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,                // Don't include params in the SQL log
			Colorful:                  false,                // Disable color
		},
	)

	db, err := gorm.Open(driver.Open(os.ExpandEnv(config.Source)), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},

		Logger: dblog,
	})
	if err != nil {
		return nil, errors.Wrap(err, "db open error")
	}
	if config.EnableMetric {
		prometheusCfg := prometheus.Config{
			DBName:          config.GetDatabaseName(),
			RefreshInterval: uint32(config.RefreshMetricInterval / time.Second),
		}
		if err = db.Use(prometheus.New(prometheusCfg)); err != nil {
			return nil, errors.Wrap(err, "db use prometheus error")
		}
	}

	register := dbresolver.Register(dbresolver.Config{})
	register = register.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	register = register.SetConnMaxLifetime(config.ConnMaxLifeTime)
	register = register.SetMaxOpenConns(config.MaxOpenConn)
	register = register.SetMaxIdleConns(config.MaxIdleConn)

	if err = db.Use(register); err != nil {
		return nil, errors.Wrap(err, "db use register error")
	}
	return newGDB(l, &config, db, driver, 0), nil
}

func NewMemoryDB(logLevel, name string) DB {
	l, err := log.NewLogger(log.FormatConsole, logLevel)
	if err != nil {
		panic(err)
	}
	config := NewDefConfig()
	config.Driver = SqliteDriver
	config.Source = fmt.Sprintf("file:%s.db?mode=memory", name)
	config.LogLevel = logLevel
	db, err := NewDB(context.Background(), l, config)
	if err != nil {
		panic(err)
	}
	return db
}

func CreateDB(logger log.Logger, config Config) error {
	driver, err := GetDriver(config.Driver)
	if err != nil {
		return errors.WithMessage(err, "get driver error")
	}
	return driver.CreateDB(logger, config)
}

func DropDB(logger log.Logger, config Config) error {
	driver, err := GetDriver(config.Driver)
	if err != nil {
		return errors.WithMessage(err, "get driver error")
	}
	return driver.DropDB(logger, config)
}

var _ DB = (*gDB)(nil)

type gDB struct {
	logger log.Logger
	config *Config

	db     *gorm.DB
	driver Driver

	rowsAffected int64
}

func (g *gDB) GetSource() string {
	return g.config.Source
}

func newGDB(logger log.Logger, config *Config, db *gorm.DB, driver Driver, rowsAffected int64) *gDB {
	return &gDB{
		logger: logger,
		config: config,

		db:     db,
		driver: driver,

		rowsAffected: rowsAffected,
	}
}

func (g *gDB) copy(db *gorm.DB) DB {
	return newGDB(g.logger, g.config, db, g.driver, g.rowsAffected)
}

func (g *gDB) WithLogger(logger log.Logger) DB {
	return newGDB(logger, g.config, g.db, g.driver, g.rowsAffected)
}

func (g *gDB) RowsAffected(number int64) DB {
	return newGDB(g.logger, g.config, g.db, g.driver, number)
}

func (g *gDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return nil
	}
	return sqlDB.Close()
}

func (g *gDB) GetDriver() Driver {
	return g.driver
}

func (g *gDB) WithContext(ctx context.Context) DB {
	return g.copy(g.db.WithContext(ctx))
}

func (g *gDB) Model(value any) DB {
	return g.copy(g.db.Model(value))
}

func (g *gDB) Where(query any, args ...any) DB {
	return g.copy(g.db.Where(query, args...))
}

func (g *gDB) Limit(limit int) DB {
	return g.copy(g.db.Limit(limit))
}

func (g *gDB) Scopes(funcs ...func(DB) DB) DB {
	fns := make([]func(db *gorm.DB) *gorm.DB, 0, len(funcs))
	for _, f := range funcs {
		fn := f
		fns = append(fns, func(db *gorm.DB) *gorm.DB { fn(g.copy(db)); return db })
	}
	return g.copy(g.db.Scopes(fns...))
}

func (g *gDB) Offset(offset int) DB {
	return g.copy(g.db.Offset(offset))
}

func (g *gDB) Order(value any) DB {
	return g.copy(g.db.Order(value))
}

func (g *gDB) Count(count *int64) DB {
	return g.copy(g.db.Count(count))
}

func (g *gDB) Group(query string) DB {
	return g.copy(g.db.Group(query))
}

func (g *gDB) Distinct(args ...any) DB {
	return g.copy(g.db.Distinct(args...))
}

func (g *gDB) Select(query any, args ...any) DB {
	return g.copy(g.db.Select(query, args...))
}

func (g *gDB) Find(dest any, conds ...any) error {
	err := g.db.Find(dest, conds...).Error
	if err != nil {
		g.logger.Error("db find error", "dest", dest, "conds", conds, "error", err)
		return errors.Wrap(err, "db find error")
	}
	return nil
}

func (g *gDB) First(dest any, conds ...any) (bool, error) {
	err := g.db.First(dest, conds...).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		g.logger.Error("db first error", "dest", dest, "conds", conds, "error", err)
		return false, errors.Wrap(err, "db first error")
	}
	return true, nil
}

func (g *gDB) MustFirst(dest any, conds ...any) error {
	return errors.Wrap(g.db.First(dest, conds...).Error, "db must first error")
}

func (g *gDB) Transaction(fn func(tx DB) error) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		return fn(g.copy(tx))
	})
}

func (g *gDB) Begin() DB {
	return g.copy(g.db.Begin())
}

func (g *gDB) Commit() error {
	return g.db.Commit().Error
}

func (g *gDB) Rollback() error {
	return g.db.Rollback().Error
}

func (g *gDB) Create(value any) error {
	tx := g.db.Create(value)
	if err := tx.Error; err != nil {
		g.logger.Error("db create error", "value", value, "error", err)
		return errors.Wrap(err, "db create error")
	}
	if (g.rowsAffected == 0 && tx.RowsAffected != 1) || (g.rowsAffected > 0 && tx.RowsAffected != g.rowsAffected) {
		g.logger.Error("db create error", "value", value, "rows affected", tx.RowsAffected)
		return errors.Errorf("db create error, rows affected: %d, expected: 1", tx.RowsAffected)
	}
	return nil
}

func (g *gDB) Update(column string, value any) error {
	tx := g.db.Update(column, value)
	if err := tx.Error; err != nil {
		g.logger.Error("db update error", "column", column, "value", value, "error", err)
		return errors.Wrap(err, "db update error")
	}
	if g.rowsAffected > 0 && tx.RowsAffected != g.rowsAffected {
		g.logger.Error("db update error", "column", column, "value", value, "rows affected", tx.RowsAffected)
		return errors.Errorf("db update error, rows affected: %d, expected: %d", tx.RowsAffected, g.rowsAffected)
	}
	return nil
}

func (g *gDB) Updates(values any) error {
	tx := g.db.Updates(values)
	if err := tx.Error; err != nil {
		g.logger.Error("db updates error", "values", values, "error", err)
		return errors.Wrap(err, "db updates error")
	}
	if g.rowsAffected > 0 && tx.RowsAffected != g.rowsAffected {
		g.logger.Error("db updates error", "values", values, "rows affected", tx.RowsAffected)
		return errors.Errorf("db updates error, rows affected: %d, expected: %d", tx.RowsAffected, g.rowsAffected)
	}
	return nil
}

func (g *gDB) Delete(value any, conds ...any) error {
	tx := g.db.Delete(value, conds...)
	if err := tx.Error; err != nil {
		g.logger.Error("db delete error", "value", value, "conds", conds, "error", err)
		return errors.Wrap(err, "db delete error")
	}
	if g.rowsAffected > 0 && tx.RowsAffected != g.rowsAffected {
		g.logger.Error("db delete error", "value", value, "conds", conds, "rows affected", tx.RowsAffected)
		return errors.Errorf("db delete error, rows affected: %d, expected: %d", tx.RowsAffected, g.rowsAffected)
	}
	return nil
}

func (g *gDB) Exec(sql string, values ...any) error {
	if err := g.db.Exec(sql, values...).Error; err != nil {
		g.logger.Error("db exec error", "sql", sql, "values", values, "error", err)
		return errors.Wrap(err, "db exec error")
	}
	return nil
}

func (g *gDB) AutoMigrate(dst ...any) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range g.driver.MigrateOptions() {
			tx = tx.Set(key, value)
		}
		if err := tx.Migrator().AutoMigrate(dst...); err != nil {
			return errors.Wrap(err, "db migration error")
		}
		return nil
	})
}
