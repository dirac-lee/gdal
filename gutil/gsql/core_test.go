package gsql

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParseTypeSlow(t *testing.T) {

	PatchConvey(t.Name(), t, func() {
		tests := []struct {
			name    string
			args    reflect.Type
			want    *sqlType
			wantErr func(string, error) bool
		}{
			{"when field is int", reflect.TypeOf(struct {
				A int `json:"a"`
			}{}), &sqlType{}, func(msg string, err error) bool {
				SoMsg(msg, err, ShouldNotBeNil)
				SoMsg(msg, err.Error(), ShouldContainSubstring, "must be pointer, but got int")
				return false
			}},
			{"when no `sql_field` tag", reflect.TypeOf(struct {
				A *int `json:"a"`
			}{}), &sqlType{}, func(msg string, err error) bool {
				SoMsg(msg, err, ShouldNotBeNil)
				SoMsg(msg, err.Error(), ShouldContainSubstring, "need sql_field tag")
				return false
			}},
			{"when invalid `sql_expr` tag", reflect.TypeOf(struct {
				A *int `sql_field:"a" sql_expr:"x"`
			}{}), &sqlType{}, func(msg string, err error) bool {
				SoMsg(msg, err, ShouldNotBeNil)
				SoMsg(msg, err.Error(), ShouldContainSubstring, "expr(x) invalid")
				return false
			}},
			{"when just 1 valid pointer is set", reflect.TypeOf(struct {
				A *int `sql_field:"a"`
			}{}), &sqlType{
				Names: []string{"A"},
				ColumnsMap: map[string]*sqlColumn{
					"A": {Name: "A", Field: "a", Kind: reflect.Pointer},
				},
			}, func(msg string, err error) bool {
				SoMsg(msg, err, ShouldBeNil)
				return true
			}},
			{"when 2 valid pointers are set", reflect.TypeOf(struct {
				A *int    `sql_field:"a"`
				B *string `sql_field:"b"`
			}{}), &sqlType{
				Names: []string{"A", "B"},
				ColumnsMap: map[string]*sqlColumn{
					"A": {Name: "A", Field: "a", Kind: reflect.Pointer},
					"B": {Name: "B", Field: "b", Kind: reflect.Pointer, Index: 1},
				},
			}, func(msg string, err error) bool {
				SoMsg(msg, err, ShouldBeNil)
				return true
			}},
		}
		for _, tt := range tests {
			PatchConvey(tt.name, func() {
				PatchConvey(tt.name+"_parseTypeSlow", func() {
					got, err := parseTypeSlow(tt.args)
					if !tt.wantErr(fmt.Sprintf("parseTypeSlow(%v)", tt.args), err) {
						return
					}
					assertSQLTypeEqual(got, tt.want)
				})
				PatchConvey(tt.name+"_parseType", func() {
					got, err := parseType(tt.args)
					if !tt.wantErr(fmt.Sprintf("parseTypeNoCache(%v)", tt.args), err) {
						return
					}
					assertSQLTypeEqual(got, tt.want)
				})
			})
		}
	})
}

func assertSQLTypeEqual(a, b *sqlType) {
	So(len(a.Names), ShouldEqual, len(b.Names))

	sort.Strings(a.Names)
	sort.Strings(b.Names)

	So(a.Names, ShouldResemble, b.Names)
	So(len(a.ColumnsMap), ShouldEqual, len(b.ColumnsMap))

	for _, name := range a.Names {
		So(a.ColumnsMap[name], ShouldNotBeNil)
		So(b.ColumnsMap[name], ShouldNotBeNil)
		So(*a.ColumnsMap[name], ShouldResemble, *b.ColumnsMap[name])
	}
}
