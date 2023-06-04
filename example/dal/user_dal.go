package dal

import (
	"github.com/dirac-lee/gdal"
	"github.com/dirac-lee/gdal/example/dal/model"
	"gorm.io/gorm"
)

type UserDAL struct {
	*gdal.GDAL[model.User, model.UserWhere, model.UserUpdate]
}

func NewUserDAL(tx *gorm.DB) *UserDAL {
	return &UserDAL{
		gdal.NewGDAL[model.User, model.UserWhere, model.UserUpdate](tx),
	}
}
