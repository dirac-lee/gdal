package gsql

import (
	"github.com/dirac-lee/gdal/gutil/gerror"
	"reflect"
)

// Implements
// @Description: 类型 T 是否实现了 Interface 接口
// @return Interface: 如果实现了，返回填入类型 T 零值的该接口；如果没实现，返回 nil
// @return bool: 是否了实现该接口
func Implements[Interface any](v any) (Interface, bool) {
	i, ok := v.(Interface)
	return i, ok
}

func GetElemValueTypeOfPtr(rv reflect.Value) (reflect.Value, reflect.Type, error) {
	rv, err := GetElemValueOfPtr(rv)
	if err != nil {
		return rv, nil, err
	}
	return rv, rv.Type(), nil
}

// GetElemValueOfPtr
//
// @Description: 获取指针底层 struct 数据类型。若非 struct 则 panic。
//
// @param rv: 底层 struct 数据类型
//
// @return reflect.Value:
//
// @example
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
// @Description: 获取指针、切片、数组底层 struct 数据类型。若非 struct 则 panic。
//
// @param rt: 底层 struct 数据类型
//
// @return reflect.Type:
//
// @example
func GetElemStructType(rt reflect.Type) (reflect.Type, error) {
	switch rt.Kind() {
	case reflect.Struct:
		return rt, nil
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return GetElemStructType(rt.Elem())
	default:
		return nil, gerror.NonStructBasedModelErr(rt)
	}
}
