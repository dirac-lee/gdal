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

// Create 创建
func (r *client) Create(ctx context.Context, po any) error {
	db := r.DB(ctx)
	if db.Error != nil {
		return db.Error
	}

	if err := db.Create(po).Error; err != nil {
		return err
	}
	return nil
}

// Delete 通过 Where 结构体 删除
//
// 返回影响的行数
func (r *client) Delete(ctx context.Context, po any, where any) (int64, error) {
	db := r.DB(ctx)
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

// Update 通过 Where + Update 结构体更新
//
// 注意：Where 结构体不能为空，即不允许不带任何条件的更新
// 返回影响的行数
func (r *client) Update(ctx context.Context, po any, where any, update any) (int64, error) {
	db := r.DB(ctx)
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

// Find .
func (r *client) Find(ctx context.Context, po any, where any, options ...QueryOption) (err error) {
	var db *gorm.DB
	db, err = r.whereDB(ctx, where, options...)
	if err != nil {
		return err
	}

	if err := db.Find(po).Error; err != nil {
		return err
	}
	return nil
}

// First 查询符合条件的第一条记录
func (r *client) First(ctx context.Context, po, where any, options ...QueryOption) error {
	db, err := r.whereDB(ctx, where, options...)
	if err != nil {
		return err
	}

	if err = db.First(po).Error; err != nil {
		return err
	}

	return nil
}

// Upsert 保存或者创建, 注意：po 必须有 ID 字段
func (r *client) Upsert(ctx context.Context, po, where, update any) (isCreated bool, err error) {
	db := r.DB(ctx)
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
	if pk == 0 {
		if err := db.Create(po).Error; err != nil {
			return false, err
		}
		isCreated = true
	} else {
		// 更新
		if _, err = r.Update(ctx, po, where, update); err != nil {
			return false, err
		}
		isCreated = false
	}
	return isCreated, nil
}

// Count 计数
func (r *client) Count(ctx context.Context, po any, where any) (int32, error) {
	db := r.DB(ctx)
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

// Exist 是否存在
func (r *client) Exist(ctx context.Context, po any, where any) (bool, error) {
	db, err := r.whereDB(ctx, where)
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

func (r *client) Save(ctx context.Context, po any) (int64, error) {
	db := r.DB(ctx)
	if db.Error != nil {
		return 0, db.Error
	}

	res := db.Save(po)
	if err := res.Error; err != nil {
		return 0, nil
	}
	return res.RowsAffected, nil
}

type idWhere struct {
	ID *int64 `sql_field:"id"`
}

// whereDB 获取 where 组装得到的 db
// 注意：如果想拿到一个 db，然后先查后改，请使用 DB 函数，不要用 Where 函数
func (r *client) whereDB(ctx context.Context, where any, options ...QueryOption) (db *gorm.DB, err error) {
	db = r.DB(ctx, options...)
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

func (r *client) DAL(makePO func() any) *DAL {
	return &DAL{
		Client: r,
		MakePO: makePO,
	}
}

// DB
// @Description: 获取 gorm.DB
// @param ctx:
// @param options:
// @return *gorm.DB:
func (r *client) DB(ctx context.Context, options ...QueryOption) *gorm.DB {
	opt := MakeQueryConfig(options)
	if r.db == nil {
		return &gorm.DB{Error: gorm.ErrInvalidTransaction}
	}
	db := r.db.WithContext(ctx)
	// TODO config
	if opt.debug {
		db = db.Debug()
	}
	if opt.readMaster {
		db = db.Clauses(dbresolver.Write)
	}
	return db
}
