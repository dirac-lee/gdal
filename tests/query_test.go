package tests_test

import (
	"github.com/dirac-lee/gdal/tests"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFind(t *testing.T) {
	Convey(t.Name(), t, func() {

		users := []tests.User{
			*GetUser("find"),
			*GetUser("find"),
			*GetUser("find"),
		}

		if err := DB.Create(&users).Error; err != nil {
			t.Fatalf("errors happened when create users: %v", err)
		}

		Convey("First", func() {
			var first tests.User
			if err := DB.Where("name = ?", "find").First(&first).Error; err != nil {
				t.Errorf("errors happened when query first: %v", err)
			} else {
				CheckUser(t, first, users[0])
			}
		})

		Convey("Last", func() {
			var last tests.User
			if err := DB.Where("name = ?", "find").Last(&last).Error; err != nil {
				t.Errorf("errors happened when query last: %v", err)
			} else {
				CheckUser(t, last, users[2])
			}
		})
	})
}
