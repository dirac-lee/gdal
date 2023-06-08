package tests_test

import (
	"github.com/dirac-lee/gdal/gutil/gptr"
	"github.com/dirac-lee/gdal/tests"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
	})
}
