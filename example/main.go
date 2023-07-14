package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm/clause"
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
		hobbies, _ := json.Marshal([]string{"swimming", "coding"})
		po := model.User{
			ID:         110,
			Name:       "dirac",
			Balance:    100,
			Hobbies:    string(hobbies),
			CreateTime: now,
			UpdateTime: now,
			Deleted:    false,
		}
		// INSERT INTO `user` (`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted`,`id`) VALUES ('dirac',100,'["swimming","coding"]','2023-07-14 21:29:08.287','2023-07-14 21:29:08.287',false,110)
		err := userDAL.Create(ctx, &po)
		fmt.Println(err)
	}

	{ // INSERT ON DUPLICATE KEY UPDATE
		now := time.Now()
		hobbies, _ := json.Marshal([]string{"cooking", "coding"})
		po := model.User{
			ID:         110,
			Name:       "dirac",
			Balance:    100,
			Hobbies:    string(hobbies),
			CreateTime: now,
			UpdateTime: now,
			Deleted:    false,
		}
		// INSERT INTO `user` (`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted`,`id`) VALUES ('dirac',100,'["cooking","coding"]','2023-07-14 21:29:08.302','2023-07-14 21:29:08.302',false,110) ON DUPLICATE KEY UPDATE `name`=VALUES(`name`),`balance`=VALUES(`balance`),`hobbies`=VALUES(`hobbies`),`create_time`=VALUES(`create_time`),`update_time`=VALUES(`update_time`),`deleted`=VALUES(`deleted`)
		err := userDAL.Clauses(clause.OnConflict{UpdateAll: true}).Create(ctx, &po)
		fmt.Println(err)
	}

	{ // multiple create
		now := time.Now()
		hobbies1, _ := json.Marshal([]string{"book", "coding"})
		hobbies2, _ := json.Marshal([]string{"book", "TV"})
		pos := []*model.User{
			{
				ID:         120,
				Name:       "bob",
				Balance:    100,
				Hobbies:    string(hobbies1),
				CreateTime: now,
				UpdateTime: now,
				Deleted:    false,
			},
			{
				ID:         130,
				Name:       "estele",
				Balance:    50,
				Hobbies:    string(hobbies2),
				CreateTime: now,
				UpdateTime: now,
				Deleted:    false,
			},
		}
		// INSERT INTO `user` (`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted`,`id`) VALUES ('bob',100,'["book","coding"]','2023-07-14 21:29:08.312','2023-07-14 21:29:08.312',false,120),('estele',50,'["book","TV"]','2023-07-14 21:29:08.312','2023-07-14 21:29:08.312',false,130)
		numCreated, err := userDAL.MCreate(ctx, &pos)
		fmt.Println(numCreated)
		fmt.Println(err)
	}

	{ // update by where condition
		where := &model.UserWhere{
			IDIn:            []int64{110, 120},
			HobbiesContains: gptr.Of("book"),
		}
		update := &model.UserUpdate{
			BalanceAdd: gptr.Of[int64](10),
		}
		// UPDATE `user` SET `balance`=balance + 10 WHERE (`id` IN (110,120) AND JSON_CONTAINS(hobbies, 'book') AND `deleted` = false)
		numUpdate, err := userDAL.MUpdate(ctx, where, update)
		fmt.Println(numUpdate)
		fmt.Println(err)
	}

	{ // update by id
		update := &model.UserUpdate{
			BalanceMinus: gptr.Of[int64](20),
		}
		// UPDATE `user` SET `balance`=balance - 20 WHERE `id` = 130
		err := userDAL.UpdateByID(ctx, 130, update)
		fmt.Println(err)
	}

	{ // general query
		var pos []*model.User
		where := &model.UserWhere{
			NameLike: gptr.Of("dirac"),
		}
		// SELECT `id`,`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted` FROM `user` WHERE (`name` LIKE '%dirac%' AND `deleted` = false)
		err := userDAL.Find(ctx, &pos, where)
		fmt.Println(err)
		fmt.Println(render.Render(pos))
	}

	{ // multiple query
		where := &model.UserWhere{
			IDIn:     []int64{110, 120},
			NameLike: gptr.Of("tel"),
		}
		// SELECT `id`,`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted` FROM `user` USE INDEX (`PRIMARY`) WHERE (`id` IN (110,120) AND `name` LIKE '%tel%' AND `deleted` = false)
		pos, err := userDAL.MQuery(ctx, where, gdal.WithMaster())
		fmt.Println(err)
		fmt.Println(render.Render(pos))
	}

	{ // query by paging: method 1
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		// SELECT count(*) FROM `user` USE INDEX (`PRIMARY`) WHERE (`id` IN (110,120) AND `deleted` = false)
		// SELECT `id`,`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted` FROM `user` USE INDEX (`PRIMARY`) WHERE (`id` IN (110,120) AND `deleted` = false) ORDER BY create_time desc LIMIT 5
		pos, total, err := userDAL.MQueryByPaging(ctx, where, gptr.Of[int64](5), nil, gptr.Of("create_time desc"))
		fmt.Println(err)
		fmt.Println(total)
		fmt.Println(render.Render(pos))
	}

	{ // query by paging: method 2
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		// SELECT count(*) FROM `user` USE INDEX (`PRIMARY`) WHERE (`id` IN (110,120) AND `deleted` = false)
		// SELECT `id`,`name`,`balance`,`hobbies`,`create_time`,`update_time`,`deleted` FROM `user` USE INDEX (`PRIMARY`) WHERE (`id` IN (110,120) AND `deleted` = false) ORDER BY create_time desc LIMIT 5
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
			// UPDATE `user` SET `balance`=balance - 20 WHERE `id` = 130
			err := userDAL.WithTx(tx).UpdateByID(ctx, 130, update)
			if err != nil {
				return err // rollback
			}

			// DELETE FROM `user` WHERE `id` = 130
			_, err = userDAL.WithTx(tx).DeleteByID(ctx, 130)
			if err != nil {
				return err // rollback
			}

			return nil // commit
		})
		fmt.Println(finalErr)
	}

	{ // physically delete
		where := &model.UserWhere{
			IDIn: []int64{110, 120},
		}
		// DELETE FROM `user` WHERE `id` IN (110,120)
		numDeleted, err := userDAL.Delete(ctx, where)
		fmt.Println(numDeleted)
		fmt.Println(err)
	}

	{ // physically delete by id
		// DELETE FROM `user` WHERE `id` = 130
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
