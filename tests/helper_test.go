package tests_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

type Config struct {
	Account   bool
	Pets      int
	Toys      int
	Company   bool
	Manager   bool
	Team      int
	Languages int
	Friends   int
	NamedPet  bool
}

func GetUser(name string, config Config) *User {
	var (
		birthday = time.Now().Round(time.Second)
		user     = User{
			Name:     name,
			Age:      18,
			Birthday: &birthday,
		}
	)

	return &user
}

func CheckUserUnscoped(t *testing.T, user User, expect User) {
	doCheckUser(t, user, expect, true)
}

func CheckUser(t *testing.T, user User, expect User) {
	doCheckUser(t, user, expect, false)
}

func doCheckUser(t *testing.T, user User, expect User, unscoped bool) {
	if user.ID != 0 {
		var newUser User
		err := db(unscoped).Where("id = ?", user.ID).First(&newUser).Error
		So(err, ShouldBeNil)

		So(newUser, ShouldResemble, user)

	}
	So(user, ShouldResemble, expect)
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
