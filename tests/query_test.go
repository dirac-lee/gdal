package tests_test

import (
	"github.com/dirac-lee/gdal/tests"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	. "gorm.io/gorm/utils/tests"
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

		var all []tests.User
		if err := DB.Where("name = ?", "find").Find(&all).Error; err != nil || len(all) != 3 {
			t.Errorf("errors happened when query find: %v, length: %v", err, len(all))
		} else {
			for idx, user := range users {
				Convey("FindAll#"+strconv.Itoa(idx+1), func() {
					CheckUser(t, all[idx], user)
				})
			}
		}

		Convey("FirstMap", func() {
			first := map[string]interface{}{}
			if err := DB.Model(&tests.User{}).Where("name = ?", "find").First(first).Error; err != nil {
				t.Errorf("errors happened when query first: %v", err)
			} else {
				for _, name := range []string{"Name", "Age", "Birthday"} {
					Convey(name, func() {
						dbName := DB.NamingStrategy.ColumnName("", name)

						switch name {
						case "Name":
							if _, ok := first[dbName].(string); !ok {
								t.Errorf("invalid data type for %v, got %#v", dbName, first[dbName])
							}
						case "Age":
							if _, ok := first[dbName].(uint); !ok {
								t.Errorf("invalid data type for %v, got %#v", dbName, first[dbName])
							}
						case "Birthday":
							if _, ok := first[dbName].(*time.Time); !ok {
								t.Errorf("invalid data type for %v, got %#v", dbName, first[dbName])
							}
						}

						reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
						AssertEqual(t, first[dbName], reflectValue.FieldByName(name).Interface())
					})
				}
			}
		})

		Convey("FirstMapWithTable", func() {
			first := map[string]interface{}{}
			if err := DB.Table("users").Where("name = ?", "find").Find(first).Error; err != nil {
				t.Errorf("errors happened when query first: %v", err)
			} else {
				for _, name := range []string{"Name", "Age", "Birthday"} {
					Convey(name, func() {
						dbName := DB.NamingStrategy.ColumnName("", name)
						resultType := reflect.ValueOf(first[dbName]).Type().Name()

						switch name {
						case "Name":
							if !strings.Contains(resultType, "string") {
								t.Errorf("invalid data type for %v, got %v %#v", dbName, resultType, first[dbName])
							}
						case "Age":
							if !strings.Contains(resultType, "int") {
								t.Errorf("invalid data type for %v, got %v %#v", dbName, resultType, first[dbName])
							}
						case "Birthday":
							if !strings.Contains(resultType, "Time") && !(DB.Dialector.Name() == "sqlite" && strings.Contains(resultType, "string")) {
								t.Errorf("invalid data type for %v, got %v %#v", dbName, resultType, first[dbName])
							}
						}

						reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
						AssertEqual(t, first[dbName], reflectValue.FieldByName(name).Interface())
					})
				}
			}
		})

		Convey("FirstPtrMap", func() {
			first := map[string]interface{}{}
			if err := DB.Model(&tests.User{}).Where("name = ?", "find").First(&first).Error; err != nil {
				t.Errorf("errors happened when query first: %v", err)
			} else {
				for _, name := range []string{"Name", "Age", "Birthday"} {
					Convey(name, func() {
						dbName := DB.NamingStrategy.ColumnName("", name)
						reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
						AssertEqual(t, first[dbName], reflectValue.FieldByName(name).Interface())
					})
				}
			}
		})

		Convey("FirstSliceOfMap", func() {
			allMap := []map[string]interface{}{}
			if err := DB.Model(&tests.User{}).Where("name = ?", "find").Find(&allMap).Error; err != nil {
				t.Errorf("errors happened when query find: %v", err)
			} else {
				for idx, user := range users {
					Convey("FindAllMap#"+strconv.Itoa(idx+1), func() {
						for _, name := range []string{"Name", "Age", "Birthday"} {
							Convey(name, func() {
								dbName := DB.NamingStrategy.ColumnName("", name)

								switch name {
								case "Name":
									if _, ok := allMap[idx][dbName].(string); !ok {
										t.Errorf("invalid data type for %v, got %#v", dbName, allMap[idx][dbName])
									}
								case "Age":
									if _, ok := allMap[idx][dbName].(uint); !ok {
										t.Errorf("invalid data type for %v, got %#v", dbName, allMap[idx][dbName])
									}
								case "Birthday":
									if _, ok := allMap[idx][dbName].(*time.Time); !ok {
										t.Errorf("invalid data type for %v, got %#v", dbName, allMap[idx][dbName])
									}
								}

								reflectValue := reflect.Indirect(reflect.ValueOf(user))
								AssertEqual(t, allMap[idx][dbName], reflectValue.FieldByName(name).Interface())
							})
						}
					})
				}
			}
		})

		Convey("FindSliceOfMapWithTable", func() {
			allMap := []map[string]interface{}{}
			if err := DB.Table("users").Where("name = ?", "find").Find(&allMap).Error; err != nil {
				t.Errorf("errors happened when query find: %v", err)
			} else {
				for idx, user := range users {
					Convey("FindAllMap#"+strconv.Itoa(idx+1), func() {
						for _, name := range []string{"Name", "Age", "Birthday"} {
							Convey(name, func() {
								dbName := DB.NamingStrategy.ColumnName("", name)
								resultType := reflect.ValueOf(allMap[idx][dbName]).Type().Name()

								switch name {
								case "Name":
									if !strings.Contains(resultType, "string") {
										t.Errorf("invalid data type for %v, got %v %#v", dbName, resultType, allMap[idx][dbName])
									}
								case "Age":
									if !strings.Contains(resultType, "int") {
										t.Errorf("invalid data type for %v, got %v %#v", dbName, resultType, allMap[idx][dbName])
									}
								case "Birthday":
									if !strings.Contains(resultType, "Time") && !(DB.Dialector.Name() == "sqlite" && strings.Contains(resultType, "string")) {
										t.Errorf("invalid data type for %v, got %v %#v", dbName, resultType, allMap[idx][dbName])
									}
								}

								reflectValue := reflect.Indirect(reflect.ValueOf(user))
								AssertEqual(t, allMap[idx][dbName], reflectValue.FieldByName(name).Interface())
							})
						}
					})
				}
			}
		})

		var models []tests.User
		if err := DB.Where("name in (?)", []string{"find"}).Find(&models).Error; err != nil || len(models) != 3 {
			t.Errorf("errors happened when query find with in clause: %v, length: %v", err, len(models))
		} else {
			for idx, user := range users {
				Convey("FindWithInClause#"+strconv.Itoa(idx+1), func() {
					CheckUser(t, models[idx], user)
				})
			}
		}

		// test array
		var models2 [3]tests.User
		if err := DB.Where("name in (?)", []string{"find"}).Find(&models2).Error; err != nil {
			t.Errorf("errors happened when query find with in clause: %v, length: %v", err, len(models2))
		} else {
			for idx, user := range users {
				Convey("array FindWithInClause#"+strconv.Itoa(idx+1), func() {
					CheckUser(t, models2[idx], user)
				})
			}
		}

		// test smaller array
		var models3 [2]tests.User
		if err := DB.Where("name in (?)", []string{"find"}).Find(&models3).Error; err != nil {
			t.Errorf("errors happened when query find with in clause: %v, length: %v", err, len(models3))
		} else {
			for idx, user := range users[:2] {
				Convey("smaller array FindWithInClause#"+strconv.Itoa(idx+1), func() {
					CheckUser(t, models3[idx], user)
				})
			}
		}

		var none []tests.User
		if err := DB.Where("name in (?)", []string{}).Find(&none).Error; err != nil || len(none) != 0 {
			t.Errorf("errors happened when query find with in clause and zero length parameter: %v, length: %v", err, len(none))
		}
	})
}
