package gdal

import (
	"context"
	"errors"
	"fmt"

	"github.com/dirac-lee/gdal/gutil/gsql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// DAL Data Access Layer.
type DAL interface {
	Create(ctx context.Context, po any) error
	Save(ctx context.Context, po any) (int64, error)
	Delete(ctx context.Context, po any, where any) (int64, error)
	Update(ctx context.Context, po any, where any, update any) (int64, error)
	Find(ctx context.Context, po any, where any, options ...QueryOption) (err error)
	First(ctx context.Context, po, where any, options ...QueryOption) error
	Count(ctx context.Context, po any, where any, options ...QueryOption) (int32, error)
	Exist(ctx context.Context, po any, where any, options ...QueryOption) (bool, error)
	DBWithCtx(ctx context.Context, options ...QueryOption) *gorm.DB
	DB(options ...QueryOption) *gorm.DB
}

// dal Data Access Layer Instance.
type dal struct {
	db *gorm.DB
}

// NewDAL new dal.
func NewDAL(tx *gorm.DB) DAL {
	cli := &dal{
		db: tx,
	}
	return cli
}

// Create record(s) of po.
//
// 💡 HINT: multiple create records when po is slice.
func (dal *dal) Create(ctx context.Context, po any) error {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return db.Error
	}

	if err := db.Create(po).Error; err != nil {
		return err
	}
	return nil
}

// Save insert when there are no conflicts, otherwise update by po's primary key.
//
// 💡 HINT: Save is suggested to be used after query then adjust some fields.
//
// ⚠️  WARNING: po must be a complete object, because Save will save all fields
// event though the field is zero value.
func (dal *dal) Save(ctx context.Context, po any) (int64, error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	res := db.Save(po)
	if err := res.Error; err != nil {
		return 0, nil
	}
	return res.RowsAffected, nil
}

// Delete delete by Where struct
func (dal *dal) Delete(ctx context.Context, po any, where any) (int64, error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	gormWhere, err := gsql.BuildSQLWhereExpr(where)
	if err != nil {
		return 0, err
	}
	if gormWhere == nil {
		return 0, fmt.Errorf("[gdal] can not delete without args")
	}
	db = db.Where(gormWhere).Delete(po) // ignore_security_alert
	return db.RowsAffected, db.Error
}

// Update updates by Where struct & Update struct. The Where struct mustn't be nil.
func (dal *dal) Update(ctx context.Context, po any, where any, update any) (int64, error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	gormWhere, err := gsql.BuildSQLWhereExpr(where)
	if err != nil {
		return 0, err
	}
	if gormWhere == nil {
		return 0, fmt.Errorf("can not update without args")
	}
	attrs, err := gsql.BuildSQLUpdate(update)
	if err != nil {
		return 0, err
	}
	if len(attrs) == 0 {
		return 0, nil
	}

	res := db.Model(po).Where(gormWhere).Updates(attrs) // ignore_security_alert
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}

// Find finds the records by Where struct
//
// 💡 HINT: options can be WithLimit, WithOffset, WithOrder, ...
//
// ⚠️  WARNING:
//
// 🚀 example:
func (dal *dal) Find(ctx context.Context, po any, where any, options ...QueryOption) (err error) {
	var db *gorm.DB
	db, err = dal.whereDB(ctx, where, options...)
	if err != nil {
		return err
	}

	if err := db.Find(po).Error; err != nil {
		return err
	}
	return nil
}

// First find the first records by Where struct
func (dal *dal) First(ctx context.Context, po, where any, options ...QueryOption) error {
	db, err := dal.whereDB(ctx, where, options...)
	if err != nil {
		return err
	}

	if err = db.First(po).Error; err != nil {
		return err
	}

	return nil
}

// Count get the count by Where struct
func (dal *dal) Count(ctx context.Context, po any, where any, options ...QueryOption) (int32, error) {
	db := dal.DBWithCtx(ctx, options...)
	if db.Error != nil {
		return 0, db.Error
	}

	gormWhere, err := gsql.BuildSQLWhereExpr(where)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(po).Where(gormWhere).Count(&count).Error; err != nil { // ignore_security_alert
		return 0, err
	}
	return int32(count), nil
}

// Exist judge if record found by where struct
func (dal *dal) Exist(ctx context.Context, po any, where any, options ...QueryOption) (bool, error) {
	db, err := dal.whereDB(ctx, where, options...)
	if err != nil {
		return false, err
	}

	if err := db.First(po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DBWithCtx embedded DB with context
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (dal *dal) DBWithCtx(ctx context.Context, options ...QueryOption) *gorm.DB {
	return dal.DB(options...).WithContext(ctx)
}

// DB embedded DB
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (dal *dal) DB(options ...QueryOption) *gorm.DB {
	opt := MakeQueryConfig(options)
	if dal.db == nil {
		return &gorm.DB{Error: gorm.ErrInvalidTransaction}
	}
	db := dal.db
	db = db.Debug()

	if opt.readMaster {
		db = db.Clauses(dbresolver.Write)
	}
	return db
}

func (dal *dal) whereDB(ctx context.Context, where any, options ...QueryOption) (db *gorm.DB, err error) {
	db = dal.DBWithCtx(ctx, options...)
	if db.Error != nil {
		return nil, db.Error
	}

	gormWhere, err := gsql.BuildSQLWhereExpr(where)
	if err != nil {
		return nil, err
	}

	opt := MakeQueryConfig(options)
	if len(opt.Selects) > 0 {
		db = db.Select(opt.Selects)
	}
	db = db.Where(gormWhere) // ignore_security_alert
	if opt.Order != nil {
		db = db.Order(*opt.Order) // ignore_security_alert
	}
	if opt.Offset != nil {
		db = db.Offset(*opt.Offset)
	}
	if opt.Limit != nil {
		db = db.Limit(*opt.Limit)
	}
	if opt.Offset != nil && opt.Limit == nil {
		return nil, fmt.Errorf("can not set offset while limit was set")
	}

	return db, nil
}
