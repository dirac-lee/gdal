package tests

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
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
