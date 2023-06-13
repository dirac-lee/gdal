package tests

import (
	"time"

	"github.com/dirac-lee/gdal/gutil/gptr"
)

type User struct {
	ID         int64      `gorm:"column:id"`
	Name       string     `gorm:"column:name"`
	Age        uint       `gorm:"column:age"`
	Birthday   *time.Time `gorm:"column:birthday"`
	CompanyID  *int       `gorm:"column:company_id"`
	ManagerID  *uint      `gorm:"column:manager_id"`
	Active     bool       `gorm:"column:active"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateTime time.Time  `gorm:"column:update_time"`
	IsDeleted  bool       `gorm:"column:is_deleted"`
}

func (u User) TableName() string {
	return "user"
}

type UserWhere struct {
	ID        *int64  `sql_field:"id"`
	Name      *string `sql_field:"name"`
	Age       *uint   `sql_field:"age"`
	CompanyID *int    `sql_field:"company_id"`
	ManagerID *uint   `sql_field:"manager_id"`
	Active    *bool   `sql_field:"active"`
	IsDeleted *bool   `sql_field:"is_deleted"`

	BirthdayGE   *time.Time `sql_field:"birthday" sql_operator:">="`
	BirthdayLT   *time.Time `sql_field:"birthday" sql_operator:"<"`
	CreateTimeGE *time.Time `sql_field:"create_time" sql_operator:">="`
	CreateTimeLT *time.Time `sql_field:"create_time" sql_operator:"<"`
	UpdateTimeGE *time.Time `sql_field:"update_time" sql_operator:">="`
	UpdateTimeLT *time.Time `sql_field:"update_time" sql_operator:"<"`

	CompanyIDIn []int  `sql_field:"company_id" sql_operator:"in"`
	ManagerIDIn []uint `sql_field:"manager_id" sql_operator:"in"`
}

func (where UserWhere) ForceIndex() string {
	if where.ID != nil {
		return "id"
	}
	return ""
}

func (where *UserWhere) InjectDefault() {
	where.IsDeleted = gptr.Of(false)
}

type UserUpdate struct {
	ID         *int64     `sql_field:"id"`
	Name       *string    `sql_field:"name"`
	Age        *uint      `sql_field:"age"`
	Birthday   *time.Time `sql_field:"birthday"`
	CompanyID  *int       `sql_field:"company_id"`
	ManagerID  *uint      `sql_field:"manager_id"`
	Active     *bool      `sql_field:"active"`
	CreateTime *time.Time `sql_field:"create_time"`
	UpdateTime *time.Time `sql_field:"update_time"`
	IsDeleted  *bool      `sql_field:"is_deleted"`
}
