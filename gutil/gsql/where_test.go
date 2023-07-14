package gsql

import (
	"github.com/dirac-lee/gdal/gutil/gslice"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils/tests"

	"github.com/dirac-lee/gdal/gutil/gptr"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildSQLWhere(t *testing.T) {

	PatchConvey(t.Name(), t, func() {
		testBuildSQLWhere := func(opt any, check func(clause.Expression, error)) {
			check(BuildSQLWhereExpr(opt))
		}

		PatchConvey("invalid sql_operator", func() {
			// sql_operator 非法
			testBuildSQLWhere(struct {
				Name *string `sql_field:"name" sql_operator:"invalid"`
			}{}, func(where clause.Expression, err error) {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, `field(Name) operator(invalid) invalid`)
			})
		})

		PatchConvey("nil field passed", func() {
			testBuildSQLWhere(struct {
				Name *string `sql_field:"name"`
			}{}, func(where clause.Expression, err error) {
				So(err, ShouldBeNil)
				query, args := buildClauses(where)
				So(query, ShouldEqual, "SELECT * FROM `users`")
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
			}, func(where clause.Expression, err error) {
				So(err, ShouldBeNil)
				query, args := buildClauses(where)
				So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` = ?")
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
			}, func(where clause.Expression, err error) {
				So(err, ShouldBeNil)
				query, args := buildClauses(where)
				So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`name_jjj` = ? AND `age_hhh` = ?)")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` IN (?,?,?) AND `name` NOT IN (?,?))")
					So(args, ShouldHaveLength, 5)
					wantArgs := []any{int64(1), int64(2), int64(3), "a", "b"}
					So(args, ShouldResemble, wantArgs)
				})
			})

			PatchConvey("*[]T AND pass ptr to empty slice", func() {
				testBuildSQLWhere(struct {
					IDs   *[]int64  `sql_field:"id" sql_operator:"in"`
					Names *[]string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   &idsEmpty,
					Names: &namesEmpty,
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` IN (?) AND `name` NOT IN (?))")
					So(args, ShouldHaveLength, 2)
					So(args[0], ShouldResemble, nil)
					So(args[1], ShouldResemble, nil)
				})
			})

			PatchConvey("when `in` OR `not in` non-zero slice", func() {
				// in 操作，不需要指针
				testBuildSQLWhere(struct {
					IDs   []int64  `sql_field:"id" sql_operator:"in"`
					Names []string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   ids,
					Names: names,
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` IN (?,?,?) AND `name` NOT IN (?,?))")
					wantArgs := []any{int64(1), int64(2), int64(3), "a", "b"}
					So(args, ShouldResemble, wantArgs)
				})
			})

			PatchConvey("when `in` OR `not in` nil slice", func() {
				// in 操作，不需要指针
				testBuildSQLWhere(struct {
					IDs   []int64  `sql_field:"id" sql_operator:"in"`
					Names []string `sql_field:"name" sql_operator:"not in"`
				}{}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users`")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when `in` OR `not in` empty slice", func() {
				testBuildSQLWhere(struct {
					IDs   []int64  `sql_field:"id" sql_operator:"in"`
					Names []string `sql_field:"name" sql_operator:"not in"`
				}{
					IDs:   []int64{},
					Names: []string{},
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users`")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when sql_operator is `=` but field is a slice", func() {
				testBuildSQLWhere(struct {
					IDs []int64 `sql_field:"id" sql_operator:"="`
				}{
					IDs: ids,
				}, func(where clause.Expression, err error) {
					ShouldNotBeNil(err)
				})
			})

			PatchConvey("when sql_operator is `full like`", func() {
				name := "name"
				testBuildSQLWhere(struct {
					NameLike *string `sql_field:"name" sql_operator:"full like"`
				}{
					NameLike: &name,
				}, func(where clause.Expression, err error) {
					query, args := buildClauses(where)
					So(err, ShouldBeNil)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` LIKE ?")
					So(args[0], ShouldEqual, "%name%")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` > ?")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` >= ?")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` = ?")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` < ?")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` <= ?")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` <> ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, 1)
				})
			})

			PatchConvey("<>", func() {
				testBuildSQLWhere(struct {
					ID   *int64 `sql_field:"id" sql_operator:"<>"`
					Name *bool  `sql_field:"name" sql_operator:"null"`
				}{
					ID: gptr.Of[int64](1),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `id` <> ?")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users`")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when it is null", func() {
				testBuildSQLWhere(struct {
					Name *bool `sql_field:"name" sql_operator:"null"`
				}{
					Name: gptr.Of(true),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` IS NULL")
					So(args, ShouldHaveLength, 0)
				})
			})

			PatchConvey("when it is not null", func() {
				testBuildSQLWhere(struct {
					Name *bool `sql_field:"name" sql_operator:"null"`
				}{
					Name: gptr.Of(false),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` IS NOT NULL")
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` LIKE ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "x")
				})
			})

			PatchConvey("left like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"left like"`
				}{
					Name: gptr.Of("x"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` LIKE ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "%x")
				})
			})

			PatchConvey("right like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"right like"`
				}{
					Name: gptr.Of("x"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` LIKE ?")
					So(args, ShouldHaveLength, 1)
					So(args[0], ShouldEqual, "x%")
				})
			})

			PatchConvey("full like", func() {
				testBuildSQLWhere(struct {
					Name *string `sql_field:"name" sql_operator:"full like"`
				}{
					Name: gptr.Of("x"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE `name` LIKE ?")
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
			}, func(where clause.Expression, err error) {
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
			}, func(where clause.Expression, err error) {
				So(err, ShouldBeNil)
				query, args := buildClauses(where)
				So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`user_id` = ? AND `parent_id` = ?)")
				So(args, ShouldResemble, []any{int64(2), int64(1)})
			})
		})

		PatchConvey("OR expression", func() {
			PatchConvey("sql_expr=$or AND no sql_field tag", func() {
				type WhereUser struct {
					UserID     *int64      `sql_field:"user_id"`
					UserName   *string     `sql_field:"user_name"`
					UserAge    *int64      `sql_field:"user_age"`
					OrClauses1 []WhereUser `sql_expr:"$or"`
					OrClauses2 []WhereUser `sql_expr:"$or"`
				}

				PatchConvey("only OR expression", func() {
					testBuildSQLWhere(WhereUser{
						OrClauses1: []WhereUser{
							{
								UserName: gptr.Of("dirac"),
							},
							{
								UserAge: gptr.Of[int64](18),
							},
						},
					}, func(where clause.Expression, err error) {
						So(err, ShouldBeNil)
						query, args := buildClauses(where) // TODO panic
						So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`user_name` = ? OR `user_age` = ?)")
						So(args, ShouldResemble, []any{"dirac", int64(18)})
					})
				})

				PatchConvey("combine OR & AND expression", func() {
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
					}, func(where clause.Expression, err error) {
						So(err, ShouldBeNil)
						query, args := buildClauses(where)
						So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`user_id` = ? AND (`user_name` = ? OR `user_age` = ?))")
						So(args, ShouldResemble, []any{int64(1), "dirac", int64(18)})
					})
				})

				PatchConvey("field of OR expression is nil", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
					}, func(where clause.Expression, err error) {
						So(err, ShouldBeNil)
						query, args := buildClauses(where)
						So(query, ShouldEqual, "SELECT * FROM `users` WHERE `user_id` = ?")
						So(args, ShouldResemble, []any{int64(1)})
					})
				})

				PatchConvey("field of OR expression is slice of empty", func() {
					testBuildSQLWhere(WhereUser{
						UserID: gptr.Of[int64](1),
						OrClauses1: []WhereUser{
							{},
							{},
						},
					}, func(where clause.Expression, err error) {
						So(err, ShouldBeNil)
						query, args := buildClauses(where)
						So(query, ShouldEqual, "SELECT * FROM `users` WHERE `user_id` = ?")
						So(args, ShouldResemble, []any{int64(1)})
					})
				})

				PatchConvey("OR expression embeddings", func() {
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
					}, func(where clause.Expression, err error) {
						So(err, ShouldBeNil)
						query, args := buildClauses(where)
						So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`user_id` = ? AND ((`user_age` = ? AND (`user_name` = ? OR `user_name` = ?)) OR `user_age` = ?))")
						So(args, ShouldResemble, []any{int64(1), int64(18), "bob", "dirac", int64(19)})
					})
				})

				PatchConvey("multiple OR expressions", func() {
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
					}, func(where clause.Expression, err error) {
						So(err, ShouldBeNil)
						query, args := buildClauses(where)
						So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`user_id` = ? AND (`user_age` = ? OR `user_age` = ?) AND (`user_name` = ? OR `user_name` = ?))")
						So(args, ShouldResemble, []any{int64(1), int64(18), int64(19), "bob", "dirac"})
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
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`user_age` = ? AND ((`user_id` = ? AND `user_name` = ?) OR (`user_id` = ? AND `user_name` = ?)))")
					So(args, ShouldResemble, []any{int64(18), int64(123), "bob", int64(234), "dirac"})
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
				}, func(where clause.Expression, err error) {
					So(err.Error(), ShouldContainSubstring, "struct field(OrClauses) with mix of sql_field(some_field) and expr($or) invalid")
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users`")
					So(args, ShouldResemble, []any(nil))
				})
			})
		})

		PatchConvey("json_contains expression", func() {
			type UserWhere struct {
				ID                   *int     `sql_field:"id" sql_operator:"="`
				FriendIDsContains    *string  `sql_field:"friend_ids" sql_operator:"json_contains"`
				FriendIDsContainsAny []string `sql_field:"friend_ids" sql_operator:"json_contains any"`
				FriendIDsContainsAll []string `sql_field:"friend_ids" sql_operator:"json_contains all"`
			}
			PatchConvey("when nil", func() {
				testBuildSQLWhere(UserWhere{}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users`")
					So(args, ShouldResemble, []any(nil))
				})
			})
			PatchConvey("when json_contains", func() {
				testBuildSQLWhere(UserWhere{
					ID:                gptr.Of(110),
					FriendIDsContains: gptr.Of("110"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` = ? AND JSON_CONTAINS(friend_ids, ?))")
					So(args, ShouldResemble, []any{110, "110"})
				})
			})
			PatchConvey("when json_contains any of multiple", func() {
				testBuildSQLWhere(UserWhere{
					ID:                   gptr.Of(110),
					FriendIDsContainsAny: gslice.Of("110", "111", "112"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` = ? AND (JSON_CONTAINS(friend_ids, ?) OR JSON_CONTAINS(friend_ids, ?) OR JSON_CONTAINS(friend_ids, ?)))")
					So(args, ShouldResemble, []any{110, "110", "111", "112"})
				})
			})
			PatchConvey("when json_contains any of one", func() {
				testBuildSQLWhere(UserWhere{
					ID:                   gptr.Of(110),
					FriendIDsContainsAny: gslice.Of("110"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` = ? AND JSON_CONTAINS(friend_ids, ?))")
					So(args, ShouldResemble, []any{110, "110"})
				})
			})
			PatchConvey("when json_contains all", func() {
				testBuildSQLWhere(UserWhere{
					ID:                   gptr.Of(110),
					FriendIDsContainsAll: gslice.Of("110", "111", "112"),
				}, func(where clause.Expression, err error) {
					So(err, ShouldBeNil)
					query, args := buildClauses(where)
					So(query, ShouldEqual, "SELECT * FROM `users` WHERE (`id` = ? AND (JSON_CONTAINS(friend_ids, ?) AND JSON_CONTAINS(friend_ids, ?) AND JSON_CONTAINS(friend_ids, ?)))")
					So(args, ShouldResemble, []any{110, "110", "111", "112"})
				})
			})
		})
	})
}

var db, _ = gorm.Open(tests.DummyDialector{}, nil)

func buildClauses(where clause.Expression) (result string, vars []any) {
	clauses := []clause.Interface{clause.Select{}, clause.From{}}
	if where != nil {
		w, ok := where.(clause.Where)
		if ok {
			clauses = append(clauses, w)
		} else {
			clauses = append(clauses, clause.Where{Exprs: []clause.Expression{where}})
		}
	}
	var (
		buildNames    []string
		buildNamesMap = map[string]bool{}
		user, _       = schema.Parse(&tests.User{}, &sync.Map{}, db.NamingStrategy)
		stmt          = gorm.Statement{DB: db, Table: user.Table, Schema: user, Clauses: map[string]clause.Clause{}}
	)

	for _, c := range clauses {
		if _, ok := buildNamesMap[c.Name()]; !ok {
			buildNames = append(buildNames, c.Name())
			buildNamesMap[c.Name()] = true
		}

		stmt.AddClause(c)
	}

	stmt.Build(buildNames...)

	return strings.TrimSpace(stmt.SQL.String()), stmt.Vars
}
