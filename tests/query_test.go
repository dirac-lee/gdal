package tests_test

import (
	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"strconv"
	"testing"
	"time"

	. "github.com/dirac-lee/gdal/tests"
)

func TestGORMFind(t *testing.T) {
	users := []User{
		*GetUser("find"),
		*GetUser("find"),
		*GetUser("find"),
	}

	PatchConvey(t.Name(), t, func() {
		err := DB.Create(&users).Error
		So(err, ShouldBeNil)

		PatchConvey("First", func() {
			var first User
			err := DB.Where("name = ?", "find").First(&first).Error
			So(err, ShouldBeNil)
			CheckUser(t, first, users[0])
		})

		PatchConvey("Last", func() {
			var last User
			err := DB.Where("name = ?", "find").Last(&last).Error
			So(err, ShouldBeNil)
			CheckUser(t, last, users[2])
		})

		var all []User
		err = DB.Where("name = ?", "find").Find(&all).Error
		So(err, ShouldBeNil)
		So(all, ShouldHaveLength, 3)

		for idx, user := range users {
			PatchConvey("FindAll#"+strconv.Itoa(idx+1), func() {
				CheckUser(t, all[idx], user)
			})
		}

		PatchConvey("FirstMap", func() {
			first := make(map[string]any)
			err := DB.Model(&User{}).Where("name = ?", "find").First(first).Error
			So(err, ShouldBeNil)

			for _, name := range []string{"Name", "Age", "Birthday"} {
				PatchConvey(name, func() {
					dbName := DB.NamingStrategy.ColumnName("", name)

					switch name {
					case "Name":
						_, ok := first[dbName].(string)
						So(ok, ShouldBeTrue)
					case "Age":
						_, ok := first[dbName].(uint)
						So(ok, ShouldBeTrue)
					case "Birthday":
						_, ok := first[dbName].(*time.Time)
						So(ok, ShouldBeTrue)
					}

					reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
					So(first[dbName], ShouldResemble, reflectValue.FieldByName(name).Interface())
				})
			}

		})

		PatchConvey("FirstMapWithTable", func() {
			first := make(map[string]any)
			err := DB.Table("users").Where("name = ?", "find").Find(first).Error
			So(err, ShouldBeNil)

			for _, name := range []string{"Name", "Age", "Birthday"} {
				PatchConvey(name, func() {
					dbName := DB.NamingStrategy.ColumnName("", name)
					resultType := reflect.ValueOf(first[dbName]).Type().Name()

					switch name {
					case "Name":
						So(resultType, ShouldContainSubstring, "string")
					case "Age":
						So(resultType, ShouldContainSubstring, "int")
					case "Birthday":
						if DB.Dialector.Name() == "sqlite" {
							So(resultType, ShouldContainSubstring, "string")
						} else {
							So(resultType, ShouldContainSubstring, "Time")
						}
					}

					reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
					So(first[dbName], ShouldResemble, reflectValue.FieldByName(name).Interface())
				})
			}

		})

		PatchConvey("FirstPtrMap", func() {
			first := make(map[string]any)
			err := DB.Model(&User{}).Where("name = ?", "find").First(&first).Error
			So(err, ShouldBeNil)
			for _, name := range []string{"Name", "Age", "Birthday"} {
				PatchConvey(name, func() {
					dbName := DB.NamingStrategy.ColumnName("", name)
					reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
					So(first[dbName], ShouldResemble, reflectValue.FieldByName(name).Interface())
				})
			}
		})

		PatchConvey("FirstSliceOfMap", func() {
			allMap := make([]map[string]any, 0)
			if err := DB.Model(&User{}).Where("name = ?", "find").Find(&allMap).Error; err != nil {
				t.Errorf("errors happened when query find: %v", err)
			} else {
				for idx, user := range users {
					PatchConvey("FindAllMap#"+strconv.Itoa(idx+1), func() {
						for _, name := range []string{"Name", "Age", "Birthday"} {
							PatchConvey(name, func() {
								dbName := DB.NamingStrategy.ColumnName("", name)

								switch name {
								case "Name":
									_, ok := allMap[idx][dbName].(string)
									So(ok, ShouldBeTrue)
								case "Age":
									_, ok := allMap[idx][dbName].(uint)
									So(ok, ShouldBeTrue)
								case "Birthday":
									_, ok := allMap[idx][dbName].(*time.Time)
									So(ok, ShouldBeTrue)
								}

								reflectValue := reflect.Indirect(reflect.ValueOf(user))
								ShouldResemble(allMap[idx][dbName], reflectValue.FieldByName(name).Interface())
							})
						}
					})
				}
			}
		})

		PatchConvey("FindSliceOfMapWithTable", func() {
			allMap := make([]map[string]any, 0)
			if err := DB.Table("users").Where("name = ?", "find").Find(&allMap).Error; err != nil {
				t.Errorf("errors happened when query find: %v", err)
			} else {
				for idx, user := range users {
					PatchConvey("FindAllMap#"+strconv.Itoa(idx+1), func() {
						for _, name := range []string{"Name", "Age", "Birthday"} {
							PatchConvey(name, func() {
								dbName := DB.NamingStrategy.ColumnName("", name)
								resultType := reflect.ValueOf(allMap[idx][dbName]).Type().Name()

								switch name {
								case "Name":
									So(resultType, ShouldContainSubstring, "string")
								case "Age":
									So(resultType, ShouldContainSubstring, "int")
								case "Birthday":
									if DB.Dialector.Name() == "sqlite" {
										So(resultType, ShouldContainSubstring, "string")
									} else {
										So(resultType, ShouldContainSubstring, "Time")
									}
								}

								reflectValue := reflect.Indirect(reflect.ValueOf(user))
								ShouldResemble(allMap[idx][dbName], reflectValue.FieldByName(name).Interface())
							})
						}
					})
				}
			}
		})
	})
}
