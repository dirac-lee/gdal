package tests_test

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/dirac-lee/gdal/tests"
	"gorm.io/gorm"
)

func GetUser(name string) *User {
	var (
		now      = time.Now()
		birthday = now.Round(time.Second)
		user     = User{
			Name:       name,
			Age:        18,
			Birthday:   &birthday,
			CreateTime: now,
			UpdateTime: now,
		}
	)
	return &user
}

func CheckUserUnscoped(user User, expect User) {
	doCheckUser(user, expect, true)
}

func CheckUser(user User, expect User) {
	doCheckUser(user, expect, false)
}

func doCheckUser(user User, expect User, unscoped bool) {
	if user.ID != 0 {
		var newUser User
		err := db(unscoped).Where("id = ?", user.ID).First(&newUser).Error
		So(err, ShouldBeNil)

		AssertObjEqual(newUser, user)
	}
	AssertObjEqual(user, expect)
}

func tidbSkip(t *testing.T, reason string) {
	if isTiDB() {
		t.Skipf("This test case skipped, because of TiDB '%s'", reason)
	}
}

func isTiDB() bool {
	return os.Getenv("GORM_DIALECT") == "tidb"
}

func db(unscoped bool) *gorm.DB {
	if unscoped {
		return DB.Unscoped()
	} else {
		return DB
	}
}
