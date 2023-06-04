package gdal

import (
	"fmt"
	"github.com/dirac-lee/gdal/gutil/greflect"
	"github.com/dirac-lee/gdal/gutil/gslice"
	"reflect"
	"strings"
	"sync"
)

var (
	type2Selector sync.Map // struct type -> selector []string
)

// GetSelector
// @Description: 读取 PO 类型底层 struct 的 gorm tag，构造 []string
// @param v:
// @return []string: selector
func GetSelector[PO any]() []string {
	var po PO
	return GetSelectorFromPOs(po)
}

// GetSelectorFromPOs
// @Description: 读取 pos 类型底层 struct 的 gorm tag，构造 []string
// @param pos:
// @return []string:
func GetSelectorFromPOs(pos any) []string {
	rt := reflect.TypeOf(pos)
	structType, err := greflect.GetElemStructType(rt)
	if err != nil {
		return nil // TODO
	}
	return getSelectorFromStructType(structType)
}

func getSelectorFromStructType(structType reflect.Type) []string {
	value, ok := type2Selector.Load(structType)
	if ok {
		return value.([]string)
	}
	return getSelectorFromStructTypeSlow(structType)
}

// getSelectorFromStructTypeSlow
// @Description: 不用缓存，读取 structType 的 gorm tag，构造 []string
// @param structType:
// @return []string:
func getSelectorFromStructTypeSlow(structType reflect.Type) []string {
	if structType.Kind() != reflect.Struct {
		panic("could not get selector from non-struct type")
	}
	var selector []string
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		gormTag := strings.TrimSpace(structField.Tag.Get("gorm"))
		tagKVs := getKVsFromTag(gormTag)
		columnName := tagKVs["column"]
		selector = append(selector, columnName)
	}
	type2Selector.Store(structType, selector)
	return selector
}

// getKVsFromTag
// @Description: 从 `k1:v1;k2:v2;...` 格式的 tag 中读取 kv map
// @param tag:
// @return map[string]string:
func getKVsFromTag(tag string) map[string]string {
	kvs := strings.Split(tag, ";")
	return gslice.ToMap(kvs, func(s string) (string, string) {
		kv := strings.Split(s, ":")
		if len(kv) != 2 {
			panic(fmt.Sprintf("tag must must be form of `k1:v1;k2:v2;...`, tag: %v", tag))
		}
		return kv[0], kv[1]
	})
}
