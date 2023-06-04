package gdal

import (
	"context"
	"errors"
	"fmt"
	"github.com/dirac-lee/gdal/gutil/gsql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"reflect"
)

// DAL
// @Description: Data Access Layer
type DAL interface {
	Create(ctx context.Context, po any) error
	Save(ctx context.Context, po any) (int64, error)
	Delete(ctx context.Context, po any, where any) (int64, error)
	Update(ctx context.Context, po any, where any, update any) (int64, error)
	Find(ctx context.Context, po any, where any, options ...QueryOption) (err error)
	First(ctx context.Context, po, where any, options ...QueryOption) error
	Count(ctx context.Context, po any, where any) (int32, error)
	Exist(ctx context.Context, po any, where any) (bool, error)
	DBWithCtx(ctx context.Context, options ...QueryOption) *gorm.DB
	DB(options ...QueryOption) *gorm.DB
}

// dal
// @Description: Data Access Layer Instance
type dal struct {
	db *gorm.DB
}

// NewDAL
//
// @Description: new dal
//
// @param db:
//
// @return DAL:
func NewDAL(tx *gorm.DB) DAL {
	cli := &dal{
		db: tx,
	}
	return cli
}

// Create
//
// @Description: create a record of po
//
// @param ctx:
// @param po: db model struct
//
// @return error:
//
// @example
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

// Save
//
// @Description: update, or insert when conflict.
//
// @param ctx:
// @param po: db model struct
//
// @return int64:
// @return error:
//
// @example
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

// Delete
//
// @Description: delete by Where struct
//
// @param ctx:
// @param po: db model struct
// @param where: db where struct
//
// @return int64: num of rows affected
// @return error:
//
// @example
func (dal *dal) Delete(ctx context.Context, po any, where any) (int64, error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	query, args, err := gsql.BuildSQLWhere(where)
	if err != nil {
		return 0, err
	}
	if len(args) == 0 {
		return 0, fmt.Errorf("[gdal] can not delete without args")
	}
	db = db.Where(query, args...).Delete(po) // ignore_security_alert
	return db.RowsAffected, db.Error
}

// Update
//
// @Description: update by Where struct & Update struct. The Where struct mustn't be nil.
//
// @param ctx:
// @param po: db model struct
// @param where: db where struct
// @param update: db update struct
//
// @return int64: num of rows affected
// @return error:
//
// @example
func (dal *dal) Update(ctx context.Context, po any, where any, update any) (int64, error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	query, args, err := gsql.BuildSQLWhere(where)
	if err != nil {
		return 0, err
	}
	if len(args) == 0 {
		return 0, fmt.Errorf("can not update without args")
	}
	attrs, err := gsql.BuildSQLUpdate(update)
	if err != nil {
		return 0, err
	}
	if len(attrs) == 0 {
		return 0, nil
	}

	res := db.Model(po).Where(query, args...).Updates(attrs) // ignore_security_alert
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}

// Find
//
// @Description: find the records by Where struct
//
// @param ctx:
// @param po: slice of db model struct to be injected by db rows
// @param where: db where struct
// @param options: db query options, e.g. WithLimit, WithOffset, WithOrder, ...
//
// @return err:
//
// @example
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

// First
//
// @Description: find the first records by Where struct
//
// @param ctx:
// @param po: db model struct to be injected by db row
// @param where: db where struct
// @param options: db query options, e.g. WithLimit, WithOffset, WithOrder, ...
//
// @return error:
//
// @example
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

// Upsert
//
// @Description: update (when exists) or insert (when absent).
//
// The db model struct should include field `ID` thus we can
// use ID field to identify the primary key and judge whether
// the record found
//
// @param ctx:
// @param po:
// @param where:
// @param update:
//
// @return isCreated:
// @return err:
//
// @example
func (dal *dal) Upsert(ctx context.Context, po, where, update any) (isCreated bool, err error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return false, db.Error
	}

	query, args, err := gsql.BuildSQLWhere(where)
	if err != nil {
		return false, err
	}

	// 先查
	if err = db.Where(query, args...).First(po).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { // ignore_security_alert
		return false, err
	}

	pov := reflect.ValueOf(po)
	if pov.Kind() == reflect.Ptr {
		pov = pov.Elem()
	}

	pk := pov.FieldByName("ID").Int()
	if pk == 0 { // ID is not set, i.e. record not found
		if err := db.Create(po).Error; err != nil {
			return false, err
		}
		isCreated = true
	} else { // record found, then update it
		if _, err = dal.Update(ctx, po, where, update); err != nil {
			return false, err
		}
		isCreated = false
	}
	return isCreated, nil
}

// Count
//
// @Description: count by Where struct
//
// @param ctx:
// @param po: db model struct
// @param where: db where struct
//
// @return int32: count
// @return error:
//
// @example
func (dal *dal) Count(ctx context.Context, po any, where any) (int32, error) {
	db := dal.DBWithCtx(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	query, args, err := gsql.BuildSQLWhere(where)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(po).Where(query, args...).Count(&count).Error; err != nil { // ignore_security_alert
		return 0, err
	}
	return int32(count), nil
}

// Exist
//
// @Description: judge if record found by where struct
//
// @param ctx:
// @param po: db model struct
// @param where: db where struct
//
// @return bool: where found
// @return error:
//
// @example
func (dal *dal) Exist(ctx context.Context, po any, where any) (bool, error) {
	db, err := dal.whereDB(ctx, where)
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

// whereDB 获取 where 组装得到的 db
// 注意：如果想拿到一个 db，然后先查后改，请使用 DB 函数，不要用 Where 函数
func (dal *dal) whereDB(ctx context.Context, where any, options ...QueryOption) (db *gorm.DB, err error) {
	db = dal.DBWithCtx(ctx, options...)
	if db.Error != nil {
		return nil, db.Error
	}

	query, args, err := gsql.BuildSQLWhere(where)
	if err != nil {
		return nil, err
	}

	opt := MakeQueryConfig(options)
	if len(opt.Selects) > 0 {
		db = db.Select(opt.Selects)
	}
	db = db.Where(query, args...) // ignore_security_alert
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

// DBWithCtx
// @Description: 获取带 context 的 gorm.DB
// @param ctx:
// @param options:
// @return *gorm.DB:
func (dal *dal) DBWithCtx(ctx context.Context, options ...QueryOption) *gorm.DB {
	return dal.DB(options...).WithContext(ctx)
}

// DB
// @Description: 获取 gorm.DB
// @param ctx:
// @param options:
// @return *gorm.DB:
func (dal *dal) DB(options ...QueryOption) *gorm.DB {
	opt := MakeQueryConfig(options)
	if dal.db == nil {
		return &gorm.DB{Error: gorm.ErrInvalidTransaction}
	}
	db := dal.db
	if opt.debug {
		db = db.Debug()
	}
	if opt.readMaster {
		db = db.Clauses(dbresolver.Write)
	}
	return db
}
