// Package gsql
// @Author liming.dirac
// @Date 2023/6/5
// @Description:
package gsql

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dirac-lee/gdal/gutil/gptr"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildSQLWhere(t *testing.T) {

	PatchConvey(t.Name(), t, func() {
		testBuildSQLWhere := func(opt interface{}, check func(query string, args []interface{}, err error)) {
			check(BuildSQLWhere(opt))
		}

		PatchConvey("invalid sql_operator", func() {
			// sql_operator 非法
			testBuildSQLWhere(struct {
				Name *string `sql_field:"name" sql_operator:"invalid"`
			}{}, func(query string, args []interface{}, err error) {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, `field(Name) operator(invalid) invalid`)
			})
		})

		PatchConvey("nil field passed", func() {
			testBuildSQLWhere(struct {
				Name *string `sql_field:"name"`
			}{}, func(query string, args []interface{}, err error) {
				So(err, ShouldBeNil)
				So(query, ShouldEqual, "")
				So(args, ShouldHaveLength, 0)
			})
		})

		PatchConvey("partially set", func() {
			name := "chen"
			testBuildSQLWhere(struct {
				Name *string `sql_field:"name"`
				Age  *int    `sql_field:"age"`
			}{
				Name: &name,
			}, func(query string, args []interface{}, err error) {
				So(err, ShouldBeNil)
				So(query, ShouldEqual, "`name` = ?")
				So(args, ShouldHaveLength, 1)
				So(args[0], ShouldEqual, "chen")
			})
		})

		PatchConvey("sql_field value is different from struct field name", func() {
			name := "chen"
			age := 20

			testBuildSQLWhere(struct {
				Name *string `sql_field:"name_jjj"`
				Age  *int    `sql_field:"age_hhh"`
			}{
				Name: &name,
				Age:  &age,
			}, func(query string, args []interface{}, err error) {
				So(err, ShouldBeNil)
				So(query, ShouldEqual, "`name_jjj` = ? and `age_hhh` = ?")
				So(args, ShouldHaveLength, 2)
				So(args[0], ShouldEqual, "chen")
				So(args[1], ShouldEqual, 20)
			})
		})

		PatchConvey("sql_operator is in", func() {
			ids := []int64{1, 2, 3}
			names := []string{"a", "b"}

			idsEmpty := make([]int64, 0)
			namesEmpty := make([]string, 0)

			PatchConvey("*[]T any all are non-zero", func() {
				testBuildSQLWhere(struct {
					IDs   *[]int64  `sql_field:"id" sql_operator:"in"`
					Names *[]string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   &ids,
					Names: &names,
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` in (?) and `name` not in (?)")
					So(args, ShouldHaveLength, 2)
					So(args[0], ShouldResemble, []int64{1, 2, 3})
					So(args[1], ShouldResemble, []string{"a", "b"})
				})
			})

			PatchConvey("*[]T and pass ptr to empty slice", func() {
				testBuildSQLWhere(struct {
					IDs   *[]int64  `sql_field:"id" sql_operator:"in"`
					Names *[]string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   &idsEmpty,
					Names: &namesEmpty,
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` in (?) and `name` not in (?)")
					So(args, ShouldHaveLength, 2)
					So(args[0], ShouldResemble, []int64{})
					So(args[1], ShouldResemble, []string{})
				})
			})

			PatchConvey("when `in` or `not in` non-zero slice", func() {
				// in 操作，不需要指针
				testBuildSQLWhere(struct {
					IDs   []int64  `sql_field:"id" sql_operator:"in"`
					Names []string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   ids,
					Names: names,
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` in (?) and `name` not in (?)")
					So(args, ShouldHaveLength, 2)
					So(args[0], ShouldResemble, []int64{1, 2, 3})
					So(args[1], ShouldResemble, []string{"a", "b"})
				})
			})

			PatchConvey("when `in` or `not in` nil slice", func() {
				// in 操作，不需要指针
				testBuildSQLWhere(struct {
					IDs   []int64  `sql_field:"id" sql_operator:"in"`
					Names []string `sql_field:"name" sql_operator:"not in"`
				}{}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when `in` or `not in` empty slice", func() {
				testBuildSQLWhere(struct {
					IDs   []int64  `sql_field:"id" sql_operator:"in"`
					Names []string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   []int64{},
					Names: []string{},
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when sql_operator is `=` but field is a slice", func() {
				testBuildSQLWhere(struct {
					IDs []int64 `sql_field:"id" sql_operator:"="`
				}{
					IDs: ids,
				}, func(query string, args []interface{}, err error) {
					ShouldNotBeNil(err)
				})
			})

			PatchConvey("when sql_operator is `full like`", func() {
				name := "name"
				testBuildSQLWhere(struct {
					NameLike *string `sql_field:"name" sql_operator:"full like"`
				}{
					NameLike: &name,
				}, func(query string, args []interface{}, err error) {
					So(query, ShouldEqual, "`name` like ?")
					So(args[0], ShouldEqual, "%name%")
					So(err, ShouldBeNil)
				})
			})
		})

		PatchConvey("math comparator", func() {
			PatchConvey(">", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:">"`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` > ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})

			PatchConvey(">=", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:">="`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(query string, args []interface{}, err error) {
					ShouldBeNil(err)
					So(query, ShouldEqual, "`id` >= ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})

			PatchConvey("=", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:"="`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` = ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})

			PatchConvey("<", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:"<"`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` < ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})

			PatchConvey("<=", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:"<="`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` <= ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})

			PatchConvey("!=", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:"!="`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`id` != ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})
		})

		PatchConvey("nullable check", func() {
			PatchConvey("when `null` tagged field is not set", func() {
				testBuildSQLWhere(struct {
					Name *bool `sql_field:"name" sql_operator:"null"`
				}{
					Name: nil,
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldBeEmpty)
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when it is null", func() {
				testBuildSQLWhere(struct {
					Name *bool `sql_field:"name" sql_operator:"null"`
				}{
					Name: gptr.Of(true),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`name` is null ")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when it is not null", func() {
				testBuildSQLWhere(struct {
					Name *bool `sql_field:"name" sql_operator:"null"`
				}{
					Name: gptr.Of(false),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`name` is not null ")
					So(args, ShouldHaveLength, 0)
				})
			})
		})

		PatchConvey("like operator", func() {
			PatchConvey("like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"like"`
				}{
					Name: gptr.Of("x"),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`name` like ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "x")
				})
			})

			PatchConvey("left like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"left like"`
				}{
					Name: gptr.Of("x"),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`name` like ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "%x")
				})
			})

			PatchConvey("right like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"right like"`
				}{
					Name: gptr.Of("x"),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`name` like ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "x%")
				})
			})

			PatchConvey("full like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"full like"`
				}{
					Name: gptr.Of("x"),
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`name` like ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "%x%")
				})
			})
		})

		PatchConvey("invalid operator", func() {
			testBuildSQLWhere(struct {
				Name *string `sql_field:"name" sql_operator:"invalid"`
			}{
				Name: gptr.Of("x"),
			}, func(query string, args []interface{}, err error) {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "field(Name) operator(invalid) invalid")
			})
		})

		PatchConvey("embedding", func() {
			type WhereUser struct {
				UserID *int64 `sql_field:"user_id"`
			}
			testBuildSQLWhere(struct {
				WhereUser
				ParentID *int64 `sql_field:"parent_id"`
			}{
				ParentID: gptr.Of[int64](1),
				WhereUser: WhereUser{
					UserID: gptr.Of[int64](2),
				},
			}, func(query string, args []interface{}, err error) {
				So(err, ShouldBeNil)
				So(query, ShouldEqual, "`user_id` = ? and `parent_id` = ?")
				So(args, ShouldResemble, []interface{}{int64(2), int64(1)})
			})
		})

		PatchConvey("or expression", func() {
			PatchConvey("sql_expr=$or and no sql_field tag", func() {
				type WhereUser struct {
					UserID     *int64      `sql_field:"user_id"`
					UserName   *string     `sql_field:"user_name"`
					UserAge    *int64      `sql_field:"user_age"`
					OrClauses1 []WhereUser `sql_expr:"$or"`
					OrClauses2 []WhereUser `sql_expr:"$or"`
				}

				PatchConvey("only or expression", func() {
					testBuildSQLWhere(WhereUser{
						OrClauses1: []WhereUser{
							{
								UserName: gptr.Of("dirac"),
							},
							{
								UserAge: gptr.Of[int64](18),
							},
						},
					}, func(query string, args []interface{}, err error) {
						So(err, ShouldBeNil)
						So(query, ShouldEqual, "(`user_name` = ? or `user_age` = ?)")
						So(args, ShouldResemble, []interface{}{"dirac", int64(18)})
					})
				})

				PatchConvey("combine or & and expression", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
						OrClauses1: []WhereUser{
							{
								UserName: gptr.Of("dirac"),
							},
							{
								UserAge: gptr.Of[int64](18),
							},
						},
					}, func(query string, args []interface{}, err error) {
						So(err, ShouldBeNil)
						So(query, ShouldEqual, "`user_id` = ? and (`user_name` = ? or `user_age` = ?)")
						So(args, ShouldResemble, []interface{}{int64(1), "dirac", int64(18)})
					})
				})

				PatchConvey("field of or expression is nil", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
					}, func(query string, args []interface{}, err error) {
						So(err, ShouldBeNil)
						So(query, ShouldEqual, "`user_id` = ?")
						So(args, ShouldResemble, []interface{}{int64(1)})
					})
				})

				PatchConvey("field of or expression is slice of empty", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
						OrClauses1: []WhereUser{
							{},
							{},
						},
					}, func(query string, args []interface{}, err error) {
						So(err, ShouldBeNil)
						So(query, ShouldEqual, "`user_id` = ?")
						So(args, ShouldResemble, []interface{}{int64(1)})
					})
				})

				PatchConvey("or expression embeddings", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
						OrClauses1: []WhereUser{
							{
								UserAge: gptr.Of[int64](18),
								OrClauses1: []WhereUser{
									{
										UserName: gptr.Of("bob"),
									},
									{
										UserName: gptr.Of("dirac"),
									},
								},
							},
							{
								UserAge: gptr.Of[int64](19),
							},
						},
					}, func(query string, args []interface{}, err error) {
						So(err, ShouldBeNil)
						So(query, ShouldEqual, "`user_id` = ? and ((`user_age` = ? and (`user_name` = ? or `user_name` = ?)) or `user_age` = ?)")
						So(args, ShouldResemble, []interface{}{int64(1), int64(18), "bob", "dirac", int64(19)})
					})
				})

				PatchConvey("multiple or expressions", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
						OrClauses1: []WhereUser{
							{
								UserAge: gptr.Of[int64](18),
							},
							{
								UserAge: gptr.Of[int64](19),
							},
						},
						OrClauses2: []WhereUser{
							{
								UserName: gptr.Of("bob"),
							},
							{
								UserName: gptr.Of("dirac"),
							},
						},
					}, func(query string, args []interface{}, err error) {
						So(err, ShouldBeNil)
						So(query, ShouldEqual, "`user_id` = ? and (`user_age` = ? or `user_age` = ?) and (`user_name` = ? or `user_name` = ?)")
						So(args, ShouldResemble, []interface{}{int64(1), int64(18), int64(19), "bob", "dirac"})
					})
				})
			})

			PatchConvey("sql_filed=-, sql_expr=$or", func() {
				type WhereUser struct {
					UserID    *int64      `sql_field:"user_id"`
					UserName  *string     `sql_field:"user_name"`
					UserAge   *int64      `sql_field:"user_age"`
					OrClauses []WhereUser `sql_field:"-" sql_expr:"$or"`
				}

				testBuildSQLWhere(WhereUser{
					UserAge: gptr.Of[int64](18),
					OrClauses: []WhereUser{
						{
							UserID:   gptr.Of[int64](123),
							UserName: gptr.Of("bob"),
						},
						{
							UserID:   gptr.Of[int64](234),
							UserName: gptr.Of("dirac"),
						},
					},
				}, func(query string, args []interface{}, err error) {
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "`user_age` = ? and ((`user_id` = ? and `user_name` = ?) or (`user_id` = ? and `user_name` = ?))")
					So(args, ShouldResemble, []interface{}{int64(18), int64(123), "bob", int64(234), "dirac"})
				})
			})

			PatchConvey("sql_filed=xxx, sql_expr=$or", func() {
				type WhereUser struct {
					UserID    *int64      `sql_field:"user_id"`
					UserName  *string     `sql_field:"user_name"`
					UserAge   *int64      `sql_field:"user_age"`
					OrClauses []WhereUser `sql_field:"some_field" sql_expr:"$or"`
				}

				testBuildSQLWhere(WhereUser{
					UserAge: gptr.Of[int64](18),
					OrClauses: []WhereUser{
						{
							UserID:   gptr.Of[int64](123),
							UserName: gptr.Of("bob"),
						},
						{
							UserID:   gptr.Of[int64](234),
							UserName: gptr.Of("dirac"),
						},
					},
				}, func(query string, args []interface{}, err error) {
					So(err.Error(), ShouldContainSubstring, "struct field(OrClauses) with mix of sql_field(some_field) and expr($or) invalid")
					So(query, ShouldEqual, "")
					So(args, ShouldResemble, []interface{}(nil))
				})
			})
		})
	})
}

func BenchmarkRTCache(b *testing.B) {
	type WhereUser struct {
		UserID    *int64      `sql_field:"user_id"`
		UserName  *string     `sql_field:"user_name"`
		UserAge   *int64      `sql_field:"user_age"`
		OrClauses []WhereUser `sql_expr:"$or"`
	}

	b.Run("with reflect.Type cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var user = WhereUser{
				UserID: gptr.Of[int64](1),
				OrClauses: []WhereUser{
					{
						UserName: gptr.Of("dirac"),
					},
					{
						UserAge: gptr.Of[int64](18),
					},
				},
			}
			rv := reflect.ValueOf(user)
			rt := rv.Type()
			getOrClauseList(rv, rt)
		}
	})

	b.Run("without reflect.Type cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var user = WhereUser{
				UserID: gptr.Of[int64](1),
				OrClauses: []WhereUser{
					{
						UserName: gptr.Of("dirac"),
					},
					{
						UserAge: gptr.Of[int64](18),
					},
				},
			}
			rv := reflect.ValueOf(user)
			rt := rv.Type()
			var orClauses []reflect.Value
			for i := 0; i < rt.NumField(); i++ {
				fieldValue := rv.Field(i)
				structField := rt.Field(i)
				sqlExpr := strings.TrimSpace(structField.Tag.Get("sql_expr"))
				if sqlExpr == "$or" {
					orClauses = append(orClauses, fieldValue)
				}
			}
		}
	})
}
