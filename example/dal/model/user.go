package model

import (
	"github.com/dirac-lee/gdal/gutil/gptr"
	"time"
)

// User db model struct, will be mapped to row of db by ORM.
type User struct {
	ID         int64     `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Balance    int64     `gorm:"column:balance"`
	Hobbies    string    `gorm:"column:hobbies"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Deleted    bool      `gorm:"column:deleted"`
}

// TableName the corresponding table name of User model struct
func (po User) TableName() string {
	return "user"
}

// UserWhere db where struct, will be mapped to SQL where condition by ORM.
type UserWhere struct {
	ID              *int64     `sql_field:"id" sql_operator:"="`
	IDIn            []int64    `sql_field:"id" sql_operator:"in"`
	Name            *string    `sql_field:"name" sql_operator:"="`
	NameLike        *string    `sql_field:"name" sql_operator:"full like"`
	HobbiesContains *string    `sql_field:"hobbies" sql_operator:"json_contains"`
	CreateTimeGT    *time.Time `sql_field:"create_time" sql_operator:">"`
	Deleted         *bool      `sql_field:"deleted" sql_operator:"="`
}

// ForceIndex if field ID or IDIn is set, clause `USE INDEX "PRIMARY"`
// will be automatically injected to the SQL.
func (where UserWhere) ForceIndex() string {
	if where.ID != nil || len(where.IDIn) > 0 {
		return "PRIMARY" // USE INDEX "PRIMARY"
	}
	return "" // no USE INDEX
}

// InjectDefault if you not set field `Deleted` of `UserWhere`,
// the where condition "deleted = false" will be automatically
// injected to the SQL.
func (where *UserWhere) InjectDefault() {
	where.Deleted = gptr.Of(false)
}

// UserUpdate db update struct, will be mapped to SQL update rule by ORM.
type UserUpdate struct {
	ID           *int64     `sql_field:"id"`
	Name         *string    `sql_field:"name"`
	Balance      *int64     `sql_field:"balance"`
	BalanceAdd   *int64     `sql_field:"balance" sql_expr:"+"`
	BalanceMinus *int64     `sql_field:"balance" sql_expr:"-"`
	UpdateTime   *time.Time `sql_field:"update_time"`
	Deleted      *bool      `sql_field:"deleted"`
}
