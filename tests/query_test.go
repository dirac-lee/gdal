package tests_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func TestFind(t *testing.T) {
	users := []User{
		*GetUser("find", Config{}),
		*GetUser("find", Config{}),
		*GetUser("find", Config{}),
	}

	if err := DB.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create users: %v", err)
	}

	t.Run("First", func(t *testing.T) {
		var first User
		if err := DB.Where("name = ?", "find").First(&first).Error; err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			CheckUser(t, first, users[0])
		}
	})

	t.Run("Last", func(t *testing.T) {
		var last User
		if err := DB.Where("name = ?", "find").Last(&last).Error; err != nil {
			t.Errorf("errors happened when query last: %v", err)
		} else {
			CheckUser(t, last, users[2])
		}
	})

	var all []User
	if err := DB.Where("name = ?", "find").Find(&all).Error; err != nil || len(all) != 3 {
		t.Errorf("errors happened when query find: %v, length: %v", err, len(all))
	} else {
		for idx, user := range users {
			t.Run("FindAll#"+strconv.Itoa(idx+1), func(t *testing.T) {
				CheckUser(t, all[idx], user)
			})
		}
	}

	t.Run("FirstMap", func(t *testing.T) {
		first := map[string]interface{}{}
		if err := DB.Model(&User{}).Where("name = ?", "find").First(first).Error; err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			for _, name := range []string{"Name", "Age", "Birthday"} {
				t.Run(name, func(t *testing.T) {
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

	t.Run("FirstMapWithTable", func(t *testing.T) {
		first := map[string]interface{}{}
		if err := DB.Table("users").Where("name = ?", "find").Find(first).Error; err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			for _, name := range []string{"Name", "Age", "Birthday"} {
				t.Run(name, func(t *testing.T) {
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

	t.Run("FirstPtrMap", func(t *testing.T) {
		first := map[string]interface{}{}
		if err := DB.Model(&User{}).Where("name = ?", "find").First(&first).Error; err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			for _, name := range []string{"Name", "Age", "Birthday"} {
				t.Run(name, func(t *testing.T) {
					dbName := DB.NamingStrategy.ColumnName("", name)
					reflectValue := reflect.Indirect(reflect.ValueOf(users[0]))
					AssertEqual(t, first[dbName], reflectValue.FieldByName(name).Interface())
				})
			}
		}
	})

	t.Run("FirstSliceOfMap", func(t *testing.T) {
		allMap := []map[string]interface{}{}
		if err := DB.Model(&User{}).Where("name = ?", "find").Find(&allMap).Error; err != nil {
			t.Errorf("errors happened when query find: %v", err)
		} else {
			for idx, user := range users {
				t.Run("FindAllMap#"+strconv.Itoa(idx+1), func(t *testing.T) {
					for _, name := range []string{"Name", "Age", "Birthday"} {
						t.Run(name, func(t *testing.T) {
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

	t.Run("FindSliceOfMapWithTable", func(t *testing.T) {
		allMap := []map[string]interface{}{}
		if err := DB.Table("users").Where("name = ?", "find").Find(&allMap).Error; err != nil {
			t.Errorf("errors happened when query find: %v", err)
		} else {
			for idx, user := range users {
				t.Run("FindAllMap#"+strconv.Itoa(idx+1), func(t *testing.T) {
					for _, name := range []string{"Name", "Age", "Birthday"} {
						t.Run(name, func(t *testing.T) {
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

	var models []User
	if err := DB.Where("name in (?)", []string{"find"}).Find(&models).Error; err != nil || len(models) != 3 {
		t.Errorf("errors happened when query find with in clause: %v, length: %v", err, len(models))
	} else {
		for idx, user := range users {
			t.Run("FindWithInClause#"+strconv.Itoa(idx+1), func(t *testing.T) {
				CheckUser(t, models[idx], user)
			})
		}
	}

	// test array
	var models2 [3]User
	if err := DB.Where("name in (?)", []string{"find"}).Find(&models2).Error; err != nil {
		t.Errorf("errors happened when query find with in clause: %v, length: %v", err, len(models2))
	} else {
		for idx, user := range users {
			t.Run("FindWithInClause#"+strconv.Itoa(idx+1), func(t *testing.T) {
				CheckUser(t, models2[idx], user)
			})
		}
	}

	// test smaller array
	var models3 [2]User
	if err := DB.Where("name in (?)", []string{"find"}).Find(&models3).Error; err != nil {
		t.Errorf("errors happened when query find with in clause: %v, length: %v", err, len(models3))
	} else {
		for idx, user := range users[:2] {
			t.Run("FindWithInClause#"+strconv.Itoa(idx+1), func(t *testing.T) {
				CheckUser(t, models3[idx], user)
			})
		}
	}

	var none []User
	if err := DB.Where("name in (?)", []string{}).Find(&none).Error; err != nil || len(none) != 0 {
		t.Errorf("errors happened when query find with in clause and zero length parameter: %v, length: %v", err, len(none))
	}
}

func TestQueryWithAssociation(t *testing.T) {
	user := *GetUser("query_with_association", Config{Account: true, Pets: 2, Toys: 1, Company: true, Manager: true, Team: 2, Languages: 1, Friends: 3})

	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create user: %v", err)
	}

	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	if err := DB.Where(&user).First(&User{}).Error; err != nil {
		t.Errorf("search with struct with association should returns no error, but got %v", err)
	}

	if err := DB.Where(user).First(&User{}).Error; err != nil {
		t.Errorf("search with struct with association should returns no error, but got %v", err)
	}
}

func TestFindInBatches(t *testing.T) {
	users := []User{
		*GetUser("find_in_batches", Config{}),
		*GetUser("find_in_batches", Config{}),
		*GetUser("find_in_batches", Config{}),
		*GetUser("find_in_batches", Config{}),
		*GetUser("find_in_batches", Config{}),
		*GetUser("find_in_batches", Config{}),
	}

	DB.Create(&users)

	var (
		results    []User
		totalBatch int
	)

	if result := DB.Table("users as u").Where("name = ?", users[0].Name).FindInBatches(&results, 2, func(tx *gorm.DB, batch int) error {
		totalBatch += batch

		if tx.RowsAffected != 2 {
			t.Errorf("Incorrect affected rows, expects: 2, got %v", tx.RowsAffected)
		}

		if len(results) != 2 {
			t.Errorf("Incorrect users length, expects: 2, got %v", len(results))
		}

		for idx := range results {
			results[idx].Name = results[idx].Name + "_new"
		}

		if err := tx.Save(results).Error; err != nil {
			t.Fatalf("failed to save users, got error %v", err)
		}

		return nil
	}); result.Error != nil || result.RowsAffected != 6 {
		t.Errorf("Failed to batch find, got error %v, rows affected: %v", result.Error, result.RowsAffected)
	}

	if totalBatch != 6 {
		t.Errorf("incorrect total batch, expects: %v, got %v", 6, totalBatch)
	}

	var count int64
	DB.Model(&User{}).Where("name = ?", "find_in_batches_new").Count(&count)
	if count != 6 {
		t.Errorf("incorrect count after update, expects: %v, got %v", 6, count)
	}
}
