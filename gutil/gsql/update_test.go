package gsql

import (
	"testing"

	. "github.com/bytedance/mockey"
	"github.com/dirac-lee/gdal/gutil/gptr"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm/clause"
)

func TestBuildSQLUpdate(t *testing.T) {
	PatchConvey(t.Name(), t, func() {
		testBuildSQLUpdate := func(opt any, check func(m map[string]any, err error)) {
			check(BuildSQLUpdate(opt))
		}

		PatchConvey("when all are zero-value", func() {
			testBuildSQLUpdate(struct {
				Name *string `sql_field:"name"`
			}{}, func(m map[string]any, err error) {
				So(err, ShouldBeNil)
				So(m, ShouldHaveLength, 0)
			})
		})

		PatchConvey("when field is partially set", func() {
			name := "chen"

			testBuildSQLUpdate(struct {
				Name *string `sql_field:"name"`
				Age  *int    `sql_field:"age"`
			}{
				Name: &name,
			}, func(m map[string]any, err error) {
				So(err, ShouldBeNil)
				So(m, ShouldHaveLength, 1)
				So(m["name"], ShouldEqual, "chen")
			})
		})

		PatchConvey("when `sql_field` value is different from struct field name", func() {
			name := "chen"
			age := 20
			testBuildSQLUpdate(struct {
				Name *string `sql_field:"name_jjj"`
				Age  *int    `sql_field:"age_hhh"`
			}{
				Name: &name,
				Age:  &age,
			}, func(m map[string]any, err error) {
				So(err, ShouldBeNil)
				So(m, ShouldHaveLength, 2)
				So(m["name_jjj"], ShouldEqual, "chen")
				So(m["age_hhh"], ShouldEqual, 20)
			})
		})

		PatchConvey("test `sql_expr` tag", func() {
			PatchConvey("+", func() {
				testBuildSQLUpdate(struct {
					Age *int `sql_field:"age" sql_expr:"+"`
				}{
					Age: gptr.Of(1),
				}, func(m map[string]any, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 1)
					So(m["age"], ShouldResemble, clause.Expr{SQL: "age + ?", Vars: []any{1}})
				})
			})

			PatchConvey("-", func() {
				testBuildSQLUpdate(struct {
					Age *int `sql_field:"age" sql_expr:"-"`
				}{
					Age: gptr.Of(1),
				}, func(m map[string]any, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 1)
					So(m["age"], ShouldResemble, clause.Expr{SQL: "age - ?", Vars: []any{1}})
				})
			})

			PatchConvey("json set", func() {
				PatchConvey("json set partial-non-zero-fields struct", func() {
					type data struct {
						A *string `json:"a"`
						B *int    `json:"b"`
						C *bool   `json:"c"`
					}
					testBuildSQLUpdate(struct {
						Data *data `sql_field:"data" sql_expr:"json_set"`
					}{
						Data: &data{
							A: gptr.Of("a"),
							C: gptr.Of(false),
						},
					}, func(m map[string]any, err error) {
						So(err, ShouldBeNil)
						So(m, ShouldHaveLength, 1)
						So(m["data"], ShouldResemble, clause.Expr{
							SQL:  "JSON_SET(`data` , '$.a', ?, '$.c', ?)",
							Vars: []any{"a", false},
						})
					})
				})

				PatchConvey("json set zero struct", func() {
					type data struct {
						A *string `json:"a"`
						B *int    `json:"b"`
						C *bool   `json:"c"`
					}
					testBuildSQLUpdate(struct {
						Data *data `sql_field:"data" sql_expr:"json_set"`
					}{
						Data: nil,
					}, func(m map[string]any, err error) {
						So(err, ShouldBeNil)
						So(m, ShouldHaveLength, 0)
					})
				})
			})

			PatchConvey("json set zero fields struct", func() {
				type data struct {
					A *string `json:"a"`
					B *int    `json:"b"`
					C *bool   `json:"c"`
				}
				testBuildSQLUpdate(struct {
					Data *data `sql_field:"data" sql_expr:"json_set"`
				}{
					Data: &data{
						A: nil,
						B: nil,
						C: nil,
					},
				}, func(m map[string]any, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 0)
				})
			})
		})

		PatchConvey("invalid `sql_expr`", func() {
			testBuildSQLUpdate(struct {
				Age *int `sql_field:"age" sql_expr:"invalid"`
			}{
				Age: gptr.Of(1),
			}, func(m map[string]any, err error) {
				So(err, ShouldNotBeNil)

				So(err.Error(), ShouldEqual, "field(Age) expr(invalid) invalid")
			})
		})
	})
}
