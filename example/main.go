package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dirac-lee/gdal"
	"github.com/dirac-lee/gdal/example/dal"
	"github.com/dirac-lee/gdal/example/dal/model"
	"github.com/dirac-lee/gdal/gutil/gptr"
	"github.com/luci/go-render/render"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

const (
	MysqlDSN = "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local"
)

func main() {
	var err error
	DB, err = gorm.Open(mysql.Open(MysqlDSN))
	if err != nil {
		panic(err)
	}

	RunMigrations()

	ctx := context.Background()
	userDAL := dal.NewUserDAL(DB)

	{ // create single record
		now := time.Now()
		po := model.User{
			ID:         110,
			Name:       "dirac",
			Balance:    100,
			CreateTime: now,
			UpdateTime: now,
			Deleted:    false,
		}
		err := userDAL.Create(ctx, &po)
		fmt.Println(err)
	}

	{ // multiple create
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
		numCreated, err := userDAL.MCreate(ctx, &pos)
		fmt.Println(numCreated)
		fmt.Println(err)
	}

	{ // update by where condition
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		update := &model.UserUpdate{
			BalanceAdd: gptr.Of[int64](10),
		}
		numUpdate, err := userDAL.MUpdate(ctx, where, update)
		fmt.Println(numUpdate)
		fmt.Println(err)
	}

	{ // update by id
		update := &model.UserUpdate{
			BalanceMinus: gptr.Of[int64](20),
		}
		err := userDAL.UpdateByID(ctx, 130, update)
		fmt.Println(err)
	}

	{ // general query
		var pos []*model.User
		where := &model.UserWhere{
			NameLike: gptr.Of("dirac"),
		}
		err := userDAL.Find(ctx, &pos, where)
		fmt.Println(err)
		fmt.Println(render.Render(pos))
	}

	{ // multiple query
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		pos, err := userDAL.MQuery(ctx, where, gdal.WithMaster())
		fmt.Println(err)
		fmt.Println(render.Render(pos))
	}

	{ // query by paging: method 1
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		pos, total, err := userDAL.MQueryByPaging(ctx, where, gptr.Of[int64](5), nil, gptr.Of("create_time desc"))
		fmt.Println(err)
		fmt.Println(total)
		fmt.Println(render.Render(pos))
	}

	{ // query by paging: method 2
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		pos, total, err := userDAL.MQueryByPagingOpt(ctx, where, gdal.WithLimit(5), gdal.WithOrder("create_time desc"))
		fmt.Println(err)
		fmt.Println(total)
		fmt.Println(render.Render(pos))
	}

	{ // transaction
		finalErr := DB.Transaction(func(tx *gorm.DB) error {

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
		fmt.Println(finalErr)
	}

	{ // 物理删除
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		numDeleted, err := userDAL.Delete(ctx, where)
		fmt.Println(numDeleted)
		fmt.Println(err)
	}

	{ // 通过ID物理删除
		numDeleted, err := userDAL.DeleteByID(ctx, 130)
		fmt.Println(numDeleted)
		fmt.Println(err)
	}
}

func RunMigrations() {
	var err error
	allModels := []interface{}{&model.User{}}

	DB.Migrator().DropTable("user_friends", "user_speaks")

	if err = DB.Migrator().DropTable(allModels...); err != nil {
		log.Printf("Failed to drop table, got error %v\n", err)
		os.Exit(1)
	}

	if err = DB.AutoMigrate(allModels...); err != nil {
		log.Printf("Failed to auto migrate, but got error %v\n", err)
		os.Exit(1)
	}

	for _, m := range allModels {
		if !DB.Migrator().HasTable(m) {
			log.Printf("Failed to create table for %#v\n", m)
			os.Exit(1)
		}
	}
}
