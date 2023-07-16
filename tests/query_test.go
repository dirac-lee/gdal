package tests_test

import (
	"github.com/dirac-lee/gdal"
	"github.com/dirac-lee/gdal/gutil/gsql"
	"gorm.io/gorm/clause"
	"testing"
	"time"

	"github.com/dirac-lee/gdal/gutil/gptr"
	"github.com/dirac-lee/gdal/tests"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFind(t *testing.T) {
	Convey(t.Name(), t, func() {

		users := []*tests.User{
			GetUser("find"),
			GetUser("find"),
			GetUser("find"),
		}

		_, err := UserDAL.MCreate(ctx, &users)
		if err != nil {
			t.Fatalf("errors happened when create users: %v", err)
		}

		Convey("QueryFirst", func() {
			var first *tests.User
			where := &tests.UserWhere{
				Name: gptr.Of("find"),
			}
			first, err = UserDAL.QueryFirst(ctx, where)
			if err != nil {
				t.Errorf("errors happened when query first: %v", err)
			}
			CheckUser(*first, *users[0])
		})

		Convey("Find", func() {
			Convey("when where is non-zero-all-field", func() {
				var users []tests.User
				where := &tests.UserWhere{
					Active:     gptr.Of(true),
					BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
					BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
				}
				err := UserDAL.Find(ctx, &users, where, gdal.WithLimit(10), gdal.WithOrder("birthday"))
				So(err, ShouldBeNil)
			})
			Convey("when where is zero-all-field", func() {
				var users []tests.User
				where := &tests.UserWhere{}
				err := UserDAL.Find(ctx, &users, where, gdal.WithLimit(10), gdal.WithOrder("birthday"))
				So(err, ShouldBeNil)
			})
		})

		Convey("Count", func() {
			where := &tests.UserWhere{
				Active:     gptr.Of(true),
				BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
				BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
			}
			count, err := UserDAL.Count(ctx, where)
			So(count, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("MQueryByPagingOpt", func() {
			where := &tests.UserWhere{
				Active:     gptr.Of(true),
				BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
				BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
			}
			users, total, err := UserDAL.MQueryByPagingOpt(ctx, where, gdal.WithLimit(10), gdal.WithOrder("birthday"))
			So(users, ShouldHaveLength, 0)
			So(total, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("QueryByID", func() {
			users, err := UserDAL.QueryByID(ctx, 123)
			So(users, ShouldHaveLength, 0)
			So(err, ShouldBeNil)
		})

		Convey("Query by v2", func() {
			var users []*tests.User
			err := UserDAL.DB().Where(gsql.BuildSQLWhereExpr(&tests.UserWhere{
				Active:     gptr.Of(true),
				BirthdayGE: gptr.Of(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)),
				BirthdayLT: gptr.Of(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)),
			})).Find(&users).Error
			So(users, ShouldHaveLength, 0)
			So(err, ShouldBeNil)
		})

		Convey("clauses", func() {
			users := []*tests.User{
				GetUser("find"),
				GetUser("find"),
				GetUser("find"),
			}
			total, err := UserDAL.Clauses(clause.OnConflict{DoNothing: true}).MCreate(ctx, &users)
			So(total, ShouldEqual, 3)
			So(err, ShouldBeNil)
		})
	})
}
