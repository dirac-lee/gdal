// Package tests
// @Author liming.dirac
// @Date 2023/6/8
// @Description:
package tests

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
)

func AssertObjEqual(r, e any, names ...string) {
	for _, name := range names {
		rv := reflect.Indirect(reflect.ValueOf(r))
		ev := reflect.Indirect(reflect.ValueOf(e))
		So(rv.IsValid(), ShouldEqual, ev.IsValid())
		got := rv.FieldByName(name).Interface()
		expect := ev.FieldByName(name).Interface()
		Convey(name, func() {
			So(got, ShouldEqual, expect)
		})
	}
}
