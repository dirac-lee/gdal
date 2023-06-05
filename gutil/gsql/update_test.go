// Package gsql
// @Author liming.dirac
// @Date 2023/6/5
// @Description:
package gsql

import (
	. "github.com/bytedance/mockey"
	"github.com/dirac-lee/gdal/gutil/gptr"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm/clause"
	"testing"
)

func TestBuildSQLUpdate(t *testing.T) {
	PatchConvey(t.Name(), t, func() {
		testBuildSQLUpdate := func(opt interface{}, check func(m map[string]interface{}, err error)) {
			check(BuildSQLUpdate(opt))
		}

		PatchConvey("when all are zero-value", func() {
			testBuildSQLUpdate(struct {
				Name *string `sql_field:"name"`
			}{}, func(m map[string]interface{}, err error) {
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
			}, func(m map[string]interface{}, err error) {
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
			}, func(m map[string]interface{}, err error) {
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
				}, func(m map[string]interface{}, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 1)
					So(m["age"], ShouldResemble, clause.Expr{SQL: "age + ?", Vars: []interface{}{1}})
				})
			})

			PatchConvey("-", func() {
				testBuildSQLUpdate(struct {
					Age *int `sql_field:"age" sql_expr:"-"`
				}{
					Age: gptr.Of(1),
				}, func(m map[string]interface{}, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 1)
					So(m["age"], ShouldResemble, clause.Expr{SQL: "age - ?", Vars: []interface{}{1}})
				})
			})

			PatchConvey("merge_json map", func() {
				m := map[string]interface{}{"a": "a", "b": 2, "c": false}
				testBuildSQLUpdate(struct {
					Data *map[string]interface{} `sql_field:"data" sql_expr:"merge_json"`
				}{
					Data: &m,
				}, func(m map[string]interface{}, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 1)
					So(m["data"], ShouldResemble, clause.Expr{
						SQL:  "CASE WHEN (`data` IS NULL OR `data` = '') THEN CAST(? AS JSON) ELSE JSON_MERGE_PATCH(`data`, CAST(? AS JSON)) END",
						Vars: []interface{}{"{\"a\":\"a\",\"b\":2,\"c\":false}", "{\"a\":\"a\",\"b\":2,\"c\":false}"},
					})
				})
			})

			PatchConvey("merge_json noe-zero struct", func() {
				type data struct {
					A string `json:"a"`
					B int    `json:"b"`
					C bool   `json:"c"`
				}
				testBuildSQLUpdate(struct {
					Data *data `sql_field:"data" sql_expr:"merge_json"`
				}{
					Data: &data{
						A: "a",
						B: 2,
						C: false,
					},
				}, func(m map[string]interface{}, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 1)
					So(m["data"], ShouldResemble, clause.Expr{
						SQL:  "CASE WHEN (`data` IS NULL OR `data` = '') THEN CAST(? AS JSON) ELSE JSON_MERGE_PATCH(`data`, CAST(? AS JSON)) END",
						Vars: []interface{}{"{\"a\":\"a\",\"b\":2,\"c\":false}", "{\"a\":\"a\",\"b\":2,\"c\":false}"},
					})
				})
			})

			PatchConvey("merge_json zero struct", func() {
				type data struct {
					A string `json:"a"`
					B int    `json:"b"`
					C bool   `json:"c"`
				}
				testBuildSQLUpdate(struct {
					Data *data `sql_field:"data" sql_expr:"merge_json"`
				}{
					Data: nil,
				}, func(m map[string]interface{}, err error) {
					So(err, ShouldBeNil)
					So(m, ShouldHaveLength, 0)
				})
			})

			PatchConvey("invalid `sql_expr`", func() {
				testBuildSQLUpdate(struct {
					Age *int `sql_field:"age" sql_expr:"invalid"`
				}{
					Age: gptr.Of(1),
				}, func(m map[string]interface{}, err error) {
					So(err, ShouldNotBeNil)

					So(err.Error(), ShouldEqual, "field(Age) expr(invalid) invalid")
				})
			})
		})
	})
}

func TestMergeJSONStructToJSONMap(t *testing.T) {
	PatchConvey(t.Name(), t, func() {
		PatchConvey("when just a int", func() {
			_, err := mergeJSONStructToJSONMap(0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "update(JSON_MERGE_PATCH) need struct type")
		})

		PatchConvey("when just a ptr to int", func() {
			_, err := mergeJSONStructToJSONMap(gptr.Of(1))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "update(JSON_MERGE_PATCH) need struct type")
		})

		PatchConvey("when all is zero-value", func() {
			m, err := mergeJSONStructToJSONMap(struct {
				Name *string `json:"name"`
			}{})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, map[string]interface{}{})
		})

		PatchConvey("when all is set", func() {
			m, err := mergeJSONStructToJSONMap(struct {
				Name *string `json:"name"`
			}{
				Name: gptr.Of("name1"),
			})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, map[string]interface{}{"name": "name1"})
		})

		PatchConvey("when partially zero-value", func() {
			m, err := mergeJSONStructToJSONMap(struct {
				Name *string `json:"name"`
				Age  *int    `json:"age"`
			}{
				Name: gptr.Of("name1"),
				Age:  nil,
			})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, map[string]interface{}{"name": "name1"})
		})

		PatchConvey("when non-ptr", func() {
			m, err := mergeJSONStructToJSONMap(struct {
				Name *string `json:"name"`
				Age  int32   `json:"age"`
			}{
				Name: gptr.Of("name1"),
				Age:  0,
			})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, map[string]interface{}{"name": "name1", "age": int32(0)})
		})
	})
}
