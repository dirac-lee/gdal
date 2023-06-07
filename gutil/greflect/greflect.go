package greflect

import (
	"github.com/dirac-lee/gdal/gutil/gerror"
	"reflect"
)

// Implements
//
// @Description: whether type `T` implements `Interface`
//
// @return Interface: return zero value of `T` if implements; otherwise, nil
//
// @return bool: whether implements
func Implements[Interface any](v any) (Interface, bool) {
	i, ok := v.(Interface)
	return i, ok
}

// GetElemValueTypeOfPtr
//
// @Description: get the element struct Value and Type if `rv` is a pointer to struct; otherwise, return `rv`'s.
//
// @param rv: pointer (maybe deep) to struct
//
// @return reflect.Value: element struct Value
//
// @return reflect.Type: element struct Type
//
// @return error: when `rv` is invalid or element is not a struct
//
// @example
//
//	u := &User{ ID: 110, Name: "Bob" }
//	rv := reflect.ValueOf(&u)
//	elemRv, elemRt, err := GetElemValueTypeOfPtr(rv)
//
// then `elemRv` is `User{ ID: 110, Name: "Bob" }`
// `elemRt` is User
// `err`    is nil
func GetElemValueTypeOfPtr(rv reflect.Value) (reflect.Value, reflect.Type, error) {
	rv, err := GetElemValueOfPtr(rv)
	if err != nil {
		return rv, nil, err
	}
	return rv, rv.Type(), nil
}

// GetElemValueOfPtr
//
// @Description: get the element struct Value if `rv` is a pointer to struct; otherwise, return `rv`'s.
//
// @param rv: pointer (maybe deep) to struct
//
// @return reflect.Value: element struct Value
//
// @example
//
//	u := &User{ ID: 110, Name: "Bob" }
//	rv := reflect.ValueOf(&u)
//	elemRv, err := GetElemValueOfPtr(rv)
//
// then `elemRv` is `User{ ID: 110, Name: "Bob" }`
// `err`    is nil
func GetElemValueOfPtr(rv reflect.Value) (reflect.Value, error) {
	if !rv.IsValid() {
		return rv, gerror.InvalidReflectValueErr(rv)
	}
	switch rv.Kind() {
	case reflect.Pointer:
		return GetElemValueOfPtr(rv.Elem())
	case reflect.Struct:
		return rv, nil
	default:
		return rv, gerror.NonStructBasedModelErr(rv.Type())
	}
}

// GetElemStructType
//
// @Description: 获取指针、切片、数组底层 struct 数据类型。
//
// @param rt: embedding of pointer (maybe deep), slice or array to struct.
//
// @return reflect.Type: element struct Type
//
// @example
//
//	u := []User{{ID: 110, Name: "Bob"}, {ID: 120, Name: "Dirac"}}
//	rv := reflect.TypeOf(&u)
//	elemRt, err := GetElemStructType(rv)
//
// then `elemRt` is `User`
// `err`    is nil
func GetElemStructType(rt reflect.Type) (reflect.Type, error) {
	type User struct {
		ID   int64
		Name string
	}
	switch rt.Kind() {
	case reflect.Struct:
		return rt, nil
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return GetElemStructType(rt.Elem())
	default:
		return nil, gerror.NonStructBasedModelErr(rt)
	}
}
