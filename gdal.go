package gdal

import (
	"context"

	"github.com/dirac-lee/gdal/gutil/gerror"
	"github.com/dirac-lee/gdal/gutil/gptr"
	"github.com/dirac-lee/gdal/gutil/gslice"
	"github.com/dirac-lee/gdal/gutil/gvalue"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// GDAL Generic Data Access Layer.
// Business DAL should embed *GDAL，and assign PO、Where and Update struct，so that you can extend
// or override the methods as you will.
//
// 💡 HINT: *Where can implement interface InjectDefaulter so that we can inject the customized
// default value into struct `where` when you query or update.
//
// 💡 HINT: Where can implement interface ForceIndexer, if so, we will force index when you query.
//
// ⚠️  WARNING: PO must implement interface Tabler and set the corresponding table name.
//
// ⚠️  WARNING: Fields of PO mapping to column must include tag `gorm:"column:{{column_name}}"`
// where `{{}}` represents placeholder, so we can know which columns you need.
//
// 🚀 example:
//
//	type User struct {
//		ID         int64      `gorm:"column:id"`
//		Name       string     `gorm:"column:name"`
//		Age        uint       `gorm:"column:age"`
//		Birthday   *time.Time `gorm:"column:birthday"`
//		CompanyID  *int       `gorm:"column:company_id"`
//		ManagerID  *uint      `gorm:"column:manager_id"`
//		Active     bool       `gorm:"column:active"`
//		CreateTime time.Time  `gorm:"column:create_time"`
//		UpdateTime time.Time  `gorm:"column:update_time"`
//		IsDeleted  bool       `gorm:"column:is_deleted"`
//	}
//
//	func (u User) TableName() string {
//		return "user"
//	}
//
//	type UserWhere struct {
//		ID        *int64  `sql_field:"id"`
//		Name      *string `sql_field:"name"`
//		Age       *uint   `sql_field:"age"`
//		CompanyID *int    `sql_field:"company_id"`
//		ManagerID *uint   `sql_field:"manager_id"`
//		Active    *bool   `sql_field:"active"`
//		IsDeleted *bool   `sql_field:"is_deleted"`
//
//		BirthdayGE   *time.Time `sql_field:"birthday" sql_operator:">="`
//		BirthdayLT   *time.Time `sql_field:"birthday" sql_operator:"<"`
//		CreateTimeGE *time.Time `sql_field:"create_time" sql_operator:">="`
//		CreateTimeLT *time.Time `sql_field:"create_time" sql_operator:"<"`
//		UpdateTimeGE *time.Time `sql_field:"update_time" sql_operator:">="`
//		UpdateTimeLT *time.Time `sql_field:"update_time" sql_operator:"<"`
//
//		CompanyIDIn []int  `sql_field:"company_id" sql_operator:"in"`
//		ManagerIDIn []uint `sql_field:"manager_id" sql_operator:"in"`
//	}
//
//	func (where UserWhere) ForceIndex() string {
//		if where.ID != nil {
//			return "id"
//		}
//		return ""
//	}
//
//	func (where *UserWhere) InjectDefault() {
//		where.IsDeleted = gptr.Of(false)
//	}
//
//	type UserUpdate struct {
//		ID         *int64     `sql_field:"id"`
//		Name       *string    `sql_field:"name"`
//		Age        *uint      `sql_field:"age"`
//		Birthday   *time.Time `sql_field:"birthday"`
//		CompanyID  *int       `sql_field:"company_id"`
//		ManagerID  *uint      `sql_field:"manager_id"`
//		Active     *bool      `sql_field:"active"`
//		CreateTime *time.Time `sql_field:"create_time"`
//		UpdateTime *time.Time `sql_field:"update_time"`
//		IsDeleted  *bool      `sql_field:"is_deleted"`
//	}
//
//	type UserDAL struct {
//		*gdal.GDAL[model.User, model.UserWhere, model.UserUpdate]
//	}
type GDAL[PO schema.Tabler, Where any, Update any] struct {
	DAL
}

// NewGDAL new GDAL
func NewGDAL[PO schema.Tabler, Where any, Update any](tx *gorm.DB) *GDAL[PO, Where, Update] {
	return &GDAL[PO, Where, Update]{
		NewDAL(tx),
	}
}

func (gdal *GDAL[PO, Where, Update]) MakePO() PO {
	return gvalue.Zero[PO]()
}

// TableName the corresponding table name
func (gdal *GDAL[PO, Where, Update]) TableName() string {
	return gdal.MakePO().TableName()
}

// Create insert a single record.
//
// 💡 HINT: the po should be a pointer so that we can inject the returning primary key.
//
// 💡 HINT: if you want to insert multiple records, use MCreate.
//
// 🚀 example:
//
//	user := tests.User{
//	Name:       "Ella",
//	Age:        17,
//	Birthday:   gptr.Of(time.Date(1999, 1, 1, 1, 0, 0, 0, time.Local)),
//	CompanyID:  gptr.Of(110),
//	ManagerID:  gptr.Of[uint](210),
//	Active:     true,
//	CreateTime: time.Now(),
//	UpdateTime: time.Now(),
//	IsDeleted:  false,
//	}
//	UserDAL.Create(ctx, &user)
//
// SQL:
// INSERT INTO `user` (`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`)
// VALUES ("Ella",17,"1999-01-01 01:00:00",110,210,true,"2023-06-11 09:38:14.483","2023-06-11 09:38:14.483",false) RETURNING `id`Ï
func (gdal *GDAL[PO, Where, Update]) Create(ctx context.Context, po *PO) error {
	return gdal.DAL.Create(ctx, po)
}

// MCreate insert multiple records.
//
// 💡 HINT: the pos should be a pointer to slice so that we can inject the returning primary keys.
//
// 🚀 example:
//
//	users := []*tests.User{
//		GetUser("find"),
//		GetUser("find"),
//		GetUser("find"),
//	}
//
//	_, err := UserDAL.MCreate(ctx, &users)
//
// SQL:
// INSERT INTO `user` (`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`)
// VALUES
// ("find",18,"2023-06-11 09:38:14",NULL,NULL,false,"2023-06-11 09:38:14.484","2023-06-11 09:38:14.484",false),
// ("find",18,"2023-06-11 09:38:14",NULL,NULL,false,"2023-06-11 09:38:14.484","2023-06-11 09:38:14.484",false),
// ("find",18,"2023-06-11 09:38:14",NULL,NULL,false,"2023-06-11 09:38:14.484","2023-06-11 09:38:14.484",false)
// RETURNING `id`
func (gdal *GDAL[PO, Where, Update]) MCreate(ctx context.Context, pos *[]*PO) (int64, error) {
	tx := gdal.DAL.DBWithCtx(ctx).Table(gdal.TableName()).CreateInBatches(pos, 100)
	return tx.RowsAffected, tx.Error
}

// Count
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
//
//	where := &tests.UserWhere{
//		Active:     gptr.Of(true),
//		BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//		BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	count, err := UserDAL.Count(ctx, where)
//
// SQL:
// SELECT count(*) FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= "1999-01-01 00:00:00" and `birthday` < "2019-01-01 00:00:00"
func (gdal *GDAL[PO, Where, Update]) Count(ctx context.Context, where *Where) (int64, error) {
	injectDefaultIfHas(where)                      // when field is not set in `where`,  insert customized default value  if customer has set it.
	indexedDAL := gdal.forceIndexIfHas(ctx, where) // force index if  it is set in `where`.
	count, err := indexedDAL.DAL.Count(ctx, gdal.MakePO(), where)
	return int64(count), err
}

// Find query by condition with paging options.
//
// 💡 HINT: this is the most primary multiple query in GDAL, on which other multiple query methods are based.
//
// ⚠️  WARNING: `pos` must be a pointer to (sub-)persistent objects.
// Fields mapping to column must include tag `gorm:"column:{{column_name}}"`
// where `{{}}` represents placeholder, so we can know which columns you need.
//
// 🚀 example:
//
//	var users []User
//	where := &UserWhere {
//		Active: gptr.Of(true),
//		BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//		BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	err := userDAL.Find(ctx, &users, where, gdal.WithLimit(10), gdal.WithOrder("birthday"))
//
// SQL:
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`
// FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= '1999-01-01 00:00:00'
// and `birthday` < '2019-01-01 00:00:00' ORDER BY birthday LIMIT 10
func (gdal *GDAL[PO, Where, Update]) Find(ctx context.Context, pos any, where any, options ...QueryOption) error {
	selector, err := GetSelectorFromPOs(pos) // 根据 PO gorm tag 确定 select 字段列表
	if err != nil {
		return err
	}
	injectDefaultIfHas(where)                      // when field is not set in `where`,  insert customized default value  if customer has set it.
	indexedDAL := gdal.forceIndexIfHas(ctx, where) // force index if  it is set in `where`.

	options = append(gslice.Of(WithSelects(selector)), options...) // as for selected columns, customer first.
	err = indexedDAL.DAL.Find(ctx, pos, where, options...)
	if gerror.IsErrRecordNotFound(err) {
		return nil
	}
	return err
}

// First query the first record by condition
//
// 💡 HINT: this is the most primary single query in GDAL, on which other single query methods are based.
//
// ⚠️  WARNING: `po` must be a pointer to a (sub-)persistent object.
// Fields mapping to column must include tag `gorm:"column:{{column_name}}"`
// where `{{}}` represents placeholder, so we can know which columns you need.
//
// 🚀 example:
//
//	var user User
//	where := &UserWhere {
//		Active: gptr.Of(true),
//		BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//		BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	err := userDAL.First(ctx, &user, where, gdal.WithOrder("birthday"))
//
// SQL:
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`
// FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= '1999-01-01 00:00:00'
// and `birthday` < '2019-01-01 00:00:00' ORDER BY birthday LIMIT 1
func (gdal *GDAL[PO, Where, Update]) First(ctx context.Context, po any, where any, options ...QueryOption) error {
	selector, err := GetSelectorFromPOs(po) // 根据 PO gorm tag 确定 select 字段列表
	if err != nil {
		return err
	}
	injectDefaultIfHas(where)                                      // when field is not set in `where`,  insert customized default value  if customer has set it.
	indexedDAL := gdal.forceIndexIfHas(ctx, where)                 // force index if  it is set in `where`.
	options = append(gslice.Of(WithSelects(selector)), options...) // as for selected columns, customer first.
	return indexedDAL.DAL.First(ctx, po, where, options...)
}

// MQuery query by condition with paging options.
//
// 💡 HINT: When you just need complete persistent objects by condition, this method is what you want.
//
// 🚀 example:
//
//	where := &UserWhere {
//		Active: gptr.Of(true),
//		BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//		BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	users, err := userDAL.MQuery(ctx, where, gdal.WithLimit(10), gdal.WithOrder("birthday"))
//
// SQL:
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`
// FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= '1999-01-01 00:00:00'
// and `birthday` < '2019-01-01 00:00:00' ORDER BY birthday LIMIT 10
func (gdal *GDAL[PO, Where, Update]) MQuery(ctx context.Context, where *Where, options ...QueryOption) ([]*PO, error) {
	var pos []*PO
	err := gdal.Find(ctx, &pos, where, options...)
	return pos, err
}

// MQueryByIDs query by primary keys.
//
// 💡 HINT: When you just need complete persistent objects by primary key list, this method is what you want.
//
// ⚠️  WARNING: nothing returns when ids is empty slice.å
//
// 🚀 example:
//
//	users, err := userDAL.MQueryByIDs(ctx, gslice.Of(123, 456, 789), gdal.WithLimit(10), gdal.WithOrder("birthday"))
//
// SQL:
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`
// FROM `user` WHERE `id` in (123, 456, 789) ORDER BY birthday LIMIT 10
func (gdal *GDAL[PO, Where, Update]) MQueryByIDs(ctx context.Context, ids []int64, options ...QueryOption) ([]*PO, error) {
	where := &idWhere{
		IDMustIn: &ids,
	}
	var pos []*PO
	err := gdal.Find(ctx, &pos, where, options...)
	return pos, err
}

// MQueryByPagingOpt query by paging options.
//
// 💡 HINT: ref Count and Find
//
// ⚠️  WARNING: the second return is the number of total records satisfy
// where condition in spite of limit and offset.
//
// 🚀 example:
//
//	where := &tests.UserWhere{
//	Active:     gptr.Of(true),
//	BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//	BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	users, total, err := UserDAL.MQueryByPagingOpt(ctx, where, gdal.WithLimit(10), gdal.WithOrder("birthday"))
//
// SQL:
// SELECT count(*) FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= "1999-01-01 00:00:00" and `birthday` < "2019-01-01 00:00:00"
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted` FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= "1999-01-01 00:00:00" and `birthday` < "2019-01-01 00:00:00" ORDER BY birthday LIMIT 10
func (gdal *GDAL[PO, Where, Update]) MQueryByPagingOpt(ctx context.Context, where *Where, options ...QueryOption) ([]*PO, int64, error) {
	count, err := gdal.Count(ctx, where)
	if err != nil || count == 0 { // skip query when count = 0
		return nil, 0, err
	}
	opt := MakeQueryConfig(options)
	if opt.Limit != nil && *opt.Limit == 0 { // skip query when limit = 0
		return nil, count, err
	}

	var pos []*PO
	pos, err = gdal.MQuery(ctx, where, options...)
	return pos, count, err
}

// MQueryByPaging query by paging.
//
// 💡 HINT: ref Count and Find
//
// ⚠️  WARNING: the second return is the number of total records satisfy
// where condition in spite of limit and offset.
//
// 🚀 example:
//
//	where := &tests.UserWhere{
//		Active:     gptr.Of(true),
//		BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//		BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	users, total, err := UserDAL.MQueryByPaging(ctx, where, gptr.Of(10), nil, gptr.Of("birthday"))
//
// SQL:
// SELECT count(*) FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= "1999-01-01 00:00:00" and `birthday` < "2019-01-01 00:00:00"
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted` FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= "1999-01-01 00:00:00" and `birthday` < "2019-01-01 00:00:00" ORDER BY birthday LIMIT 10
func (gdal *GDAL[PO, Where, Update]) MQueryByPaging(ctx context.Context, where *Where, limit *int64, offset *int64, order *string) ([]*PO, int64, error) {
	options := buildQueryOptions(limit, offset, order)
	return gdal.MQueryByPagingOpt(ctx, where, options...)
}

// QueryFirst query the first record by condition.
//
// 💡 HINT: ref First.
//
// 🚀 example:
//
//	var user User
//	where := &UserWhere {
//		Active: gptr.Of(true),
//		BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
//		BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
//	}
//	user, err := userDAL.QueryFirst(ctx, where, gdal.WithOrder("birthday"))
//
// SQL:
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted`
// FROM `user` WHERE `active` = true and `is_deleted` = false and `birthday` >= '1999-01-01 00:00:00'
// and `birthday` < '2019-01-01 00:00:00' ORDER BY birthday LIMIT 1
func (gdal *GDAL[PO, Where, Update]) QueryFirst(ctx context.Context, where *Where, options ...QueryOption) (*PO, error) {
	var po PO
	err := gdal.First(ctx, &po, where, options...)
	if err != nil {
		if gerror.IsErrRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

// QueryByID query the record by primary key
//
// ⚠️  WARNING: if primary key is not exist, return nil pointer.
//
// 🚀 example:
//
//	users, err := UserDAL.QueryByID(ctx, 123)
//
// SQL:
// SELECT `id`,`name`,`age`,`birthday`,`company_id`,`manager_id`,`active`,`create_time`,`update_time`,`is_deleted` FROM `user` WHERE `id` = 123 ORDER BY `user`.`id` LIMIT 1
func (gdal *GDAL[PO, Where, Update]) QueryByID(ctx context.Context, id int64) (*PO, error) {
	where := &idWhere{
		ID: gptr.Of(id),
	}
	var po PO
	err := gdal.First(ctx, &po, where)
	if err != nil {
		if gerror.IsErrRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &po, err
}

// MUpdate updates multiple records by condition, return success count
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) MUpdate(ctx context.Context, where *Where, update *Update) (int64, error) {
	injectDefaultIfHas(where) // when field is not set in `where`,  insert customized default value  if customer has set it.
	return gdal.DAL.Update(ctx, gdal.MakePO(), where, update)
}

// Update updates records by condition
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) Update(ctx context.Context, where *Where, update *Update) error {
	injectDefaultIfHas(where) // when field is not set in `where`,  insert customized default value  if customer has set it.
	_, err := gdal.MUpdate(ctx, where, update)
	return err
}

// UpdateByID updates single record by primary key
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) UpdateByID(ctx context.Context, id int64, update *Update) error {
	_, err := gdal.DAL.Update(ctx, gdal.MakePO(), idWhere{ID: &id}, update)
	return err
}

// Save saves single record
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) Save(ctx context.Context, po *PO) error {
	_, err := gdal.DAL.Save(ctx, po)
	return err
}

// MSave saves multiple records
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) MSave(ctx context.Context, pos *[]*PO) error {
	_, err := gdal.DAL.Save(ctx, pos)
	return err
}

// Delete deletes physically by condition
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) Delete(ctx context.Context, where *Where) (int64, error) {
	return gdal.DAL.Delete(ctx, gdal.MakePO(), where)
}

// DeleteByID deletes physically by primary key
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) DeleteByID(ctx context.Context, id int64) (int64, error) {
	return gdal.DAL.Delete(ctx, gdal.MakePO(), idWhere{ID: &id})
}

// WithTx generate a new GDAL with tx embedded
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) WithTx(tx *gorm.DB) *GDAL[PO, Where, Update] {
	return NewGDAL[PO, Where, Update](tx)
}

// DBWithCtx get embedded DB with context
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) DBWithCtx(ctx context.Context, options ...QueryOption) *gorm.DB {
	return gdal.DAL.DBWithCtx(ctx, options...)
}

// DB get embedded DB
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) DB(options ...QueryOption) *gorm.DB {
	return gdal.DAL.DB(options...)
}

// Clauses generate a new GDAL with clause supplemented.
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func (gdal *GDAL[PO, Where, Update]) Clauses(conds ...clause.Expression) *GDAL[PO, Where, Update] {
	tx := gdal.DB().Clauses(conds...)
	return NewGDAL[PO, Where, Update](tx)
}

func buildQueryOptions(limit *int64, offset *int64, order *string) []QueryOption {
	var options []QueryOption
	if limit != nil {
		options = append(options, WithLimit(int(*limit)))
	}
	if offset != nil {
		options = append(options, WithOffset(int(*offset)))
	}
	if order != nil {
		options = append(options, WithOrder(*order))
	}
	return options
}

type idWhere struct {
	ID       *int64   `sql_field:"id" sql_operator:"="`
	IDMustIn *[]int64 `sql_field:"id" sql_operator:"in"`
}

func (where *idWhere) ForceIndex() string {
	return "PRIMARY"
}
