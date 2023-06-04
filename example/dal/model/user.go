// Package model
// @Author liming.dirac
// @Date 2023/6/4
// @Description:
package model

import "time"

// User db model struct, will be mapped to row of db by ORM.
type User struct {
	ID         int64     `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Balance    int64     `gorm:"column:balance"`
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
	ID           *int64     `sql_field:"id" sql_operator:"="`
	IDIn         []int64    `sql_field:"id" sql_operator:"in"`
	Name         *string    `sql_field:"name" sql_operator:"="`
	NameLike     *string    `sql_field:"name" sql_operator:"like"`
	CreateTimeGT *time.Time `sql_field:"create_time" sql_operator:">"`
	Deleted      *bool      `sql_field:"deleted" sql_operator:"="`
}

type UserUpdate struct {
	ID           *int64     `sql_field:"id"`
	Name         *string    `sql_field:"name"`
	Balance      *int64     `gorm:"column:balance"`
	BalanceAdd   *int64     `gorm:"column:balance" sql_expr:"+"`
	BalanceMinus *int64     `gorm:"column:balance" sql_expr:"-"`
	UpdateTime   *time.Time `sql_field:"update_time"`
	Deleted      *bool      `sql_field:"deleted"`
}
