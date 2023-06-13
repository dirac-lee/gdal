package tests

import (
	"reflect"

	. "github.com/smartystreets/goconvey/convey"
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
