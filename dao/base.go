package dao

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm/schema"

	"github.com/pundiai/go-sdk/db"
)

type (
	ctxKeyTx    struct{}
	ctxKeyHasTx struct{}
)

var (
	keyTx    = ctxKeyTx{}
	keyHasTx = ctxKeyHasTx{}
)

type Model interface {
	schema.Tabler
}

type BaseDao struct {
	db    db.DB
	model Model
}

func NewDao(db db.DB, model Model) *BaseDao {
	return &BaseDao{db: db, model: model}
}

func (d *BaseDao) GetDB() db.DB {
	return d.db
}

func (d *BaseDao) Insert(model Model) error {
	return d.db.Create(model)
}

func (d *BaseDao) Count(funcs ...func(db db.DB) db.DB) (count int64) {
	d.db.Model(d.model).Scopes(funcs...).Count(&count)
	return count
}

func (d *BaseDao) UpdatesByID(id uint, data Model) error {
	err := d.db.Model(d.model).
		Where("id = ?", id).
		RowsAffected(1).
		Updates(data)
	if err != nil {
		return errors.WithMessagef(err, "id: %d", id)
	}
	return nil
}

func (d *BaseDao) DeleteByID(id uint) error {
	err := d.db.Model(d.model).
		Where("id = ?", id).
		RowsAffected(1).
		Delete(nil)
	if err != nil {
		return errors.WithMessagef(err, "id: %d", id)
	}
	return nil
}

func (d *BaseDao) GetByID(id uint, result Model) (bool, error) {
	found, err := d.db.Model(d.model).
		Where("id = ?", id).
		First(result)
	if err != nil {
		return false, errors.WithMessagef(err, "id: %d", id)
	}
	return found, nil
}

func (d *BaseDao) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.UnwrapContextDBOrDefault(ctx).Transaction(func(tx db.DB) error {
		return fn(context.WithValue(ctx, keyTx, tx))
	})
}

func (d *BaseDao) BeginTx(ctx context.Context) context.Context {
	if _, ok := ctx.Value(keyTx).(db.DB); ok {
		ctx = context.WithValue(ctx, keyHasTx, true)
		return ctx
	}
	return context.WithValue(ctx, keyTx, d.db.Begin())
}

func (*BaseDao) CommitTx(ctx context.Context) error {
	tx, canOperator := hasOperatorTx(ctx)
	if !canOperator {
		return nil
	}
	return tx.Commit()
}

func (*BaseDao) RollbackTx(ctx context.Context) error {
	tx, canOperator := hasOperatorTx(ctx)
	if !canOperator {
		return nil
	}
	return tx.Rollback()
}

func hasOperatorTx(ctx context.Context) (db.DB, bool) {
	tx, ok := ctx.Value(keyTx).(db.DB)
	if !ok {
		return nil, false
	}
	if hasTx, ok := ctx.Value(keyHasTx).(bool); ok && hasTx {
		return nil, false
	}
	return tx, true
}

func (d *BaseDao) UnwrapContextDBOrDefault(ctx context.Context) db.DB {
	value, ok := ctx.Value(keyTx).(db.DB)
	if ok {
		return value
	}
	return d.db
}

func (d *BaseDao) InsertWithCtx(ctx context.Context, model Model) error {
	return d.UnwrapContextDBOrDefault(ctx).Create(model)
}

func (d *BaseDao) UpdatesByIDWithCtx(ctx context.Context, id uint, data Model) error {
	err := d.UnwrapContextDBOrDefault(ctx).Model(d.model).
		Where("id = ?", id).
		RowsAffected(1).
		Updates(data)
	if err != nil {
		return errors.WithMessagef(err, "id: %d", id)
	}
	return nil
}
