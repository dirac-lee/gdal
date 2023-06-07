package tests_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"strconv"
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

	if config.Account {
		user.Account = Account{Number: name + "_account"}
	}

	for i := 0; i < config.Pets; i++ {
		user.Pets = append(user.Pets, &Pet{Name: name + "_pet_" + strconv.Itoa(i+1)})
	}

	for i := 0; i < config.Toys; i++ {
		user.Toys = append(user.Toys, Toy{Name: name + "_toy_" + strconv.Itoa(i+1)})
	}

	if config.Company {
		user.Company = Company{Name: "company-" + name}
	}

	if config.Manager {
		user.Manager = GetUser(name+"_manager", Config{})
	}

	for i := 0; i < config.Team; i++ {
		user.Team = append(user.Team, *GetUser(name+"_team_"+strconv.Itoa(i+1), Config{}))
	}

	for i := 0; i < config.Languages; i++ {
		name := name + "_locale_" + strconv.Itoa(i+1)
		language := Language{Code: name, Name: name}
		user.Languages = append(user.Languages, language)
	}

	for i := 0; i < config.Friends; i++ {
		user.Friends = append(user.Friends, GetUser(name+"_friend_"+strconv.Itoa(i+1), Config{}))
	}

	if config.NamedPet {
		user.NamedPet = &Pet{Name: name + "_namepet"}
	}

	return &user
}

func CheckPetUnscoped(t *testing.T, pet Pet, expect Pet) {
	doCheckPet(t, pet, expect, true)
}

func CheckPet(t *testing.T, pet Pet, expect Pet) {
	doCheckPet(t, pet, expect, false)
}

func doCheckPet(t *testing.T, pet Pet, expect Pet, unscoped bool) {
	if pet.ID != 0 {
		var newPet Pet
		if err := db(unscoped).Where("id = ?", pet.ID).First(&newPet).Error; err != nil {
			t.Fatalf("errors happened when query: %v", err)
		} else {
			AssertObjEqual(t, newPet, pet, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "UserID", "Name")
			AssertObjEqual(t, newPet, expect, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "UserID", "Name")
		}
	}

	AssertObjEqual(t, pet, expect, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "UserID", "Name")

	AssertObjEqual(t, pet.Toy, expect.Toy, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Name", "OwnerID", "OwnerType")

	if expect.Toy.Name != "" && expect.Toy.OwnerType != "pets" {
		t.Errorf("toys's OwnerType, expect: %v, got %v", "pets", expect.Toy.OwnerType)
	}
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
