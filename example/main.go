package main

import (
	"context"
	"github.com/dirac-lee/gdal"
	"github.com/dirac-lee/gdal/example/dal"
	"github.com/dirac-lee/gdal/example/dal/model"
	"github.com/dirac-lee/gdal/gutil/gptr"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

const (
	DemoDSN = "demo"
)

func main() {
	db, err := gorm.Open(mysql.Open(DemoDSN))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	userDAL := dal.NewUserDAL(db)

	{ //创建单条记录
		now := time.Now()
		po := model.User{
			ID:         110,
			Name:       "dirac",
			Balance:    100,
			CreateTime: now,
			UpdateTime: now,
			Deleted:    false,
		}
		userDAL.Create(ctx, &po)
	}

	{ // 创建多条记录
		now := time.Now()
		pos := []*model.User{
			{
				ID:         120,
				Name:       "bob",
				Balance:    100,
				CreateTime: now,
				UpdateTime: now,
				Deleted:    false,
			},
			{
				ID:         130,
				Name:       "estele",
				Balance:    50,
				CreateTime: now,
				UpdateTime: now,
				Deleted:    false,
			},
		}
		userDAL.MCreate(ctx, &pos)
	}

	{ // 物理删除
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		userDAL.Delete(ctx, where)
	}

	{ // 通过ID物理删除
		userDAL.DeleteByID(ctx, 130)
	}

	{ // 更新
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		update := &model.UserUpdate{
			BalanceAdd: gptr.Of[int64](10),
		}
		userDAL.Update(ctx, where, update)
	}

	{ // 通过 ID 更新
		update := &model.UserUpdate{
			BalanceMinus: gptr.Of[int64](20),
		}
		userDAL.UpdateByID(ctx, 130, update)
	}

	{ // 通用查询
		var pos []*model.User
		where := &model.UserWhere{
			NameLike: gptr.Of("dirac"),
		}
		userDAL.Find(ctx, &pos, where)
	}

	{ // 查询多条记录
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		pos, err := userDAL.MQuery(ctx, where)
		println(pos, err)
	}

	{ // 分页查询1
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		pos, total, err := userDAL.MQueryByPaging(ctx, where, gptr.Of[int64](5), nil, gptr.Of("create_time desc"))
		println(pos, total, err)
	}

	{ // 分页查询2
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		pos, total, err := userDAL.MQueryByPagingOpt(ctx, where, gdal.WithLimit(5), gdal.WithOrder("create_time desc"))
		println(pos, total, err)
	}

	{
		finalErr := db.Transaction(func(tx *gorm.DB) error {

			update := &model.UserUpdate{
				BalanceMinus: gptr.Of[int64](20),
			}
			err := userDAL.WithTx(tx).UpdateByID(ctx, 130, update)
			if err != nil {
				return err // rollback
			}

			_, err = userDAL.WithTx(tx).DeleteByID(ctx, 130)
			if err != nil {
				return err // rollback
			}

			return nil // commit
		})
		println(finalErr)
	}
}
