package gdal

import (
	"github.com/dirac-lee/gdal/gutil/gerror"
	"github.com/dirac-lee/gdal/gutil/greflect"
	"reflect"
	"strings"
	"sync"
)

var (
	type2Selector sync.Map // struct type -> selector []string
)

// GetSelectorFromPOs
// @Description: 读取 pos 类型底层 struct 的 gorm tag，构造 []string
// @param pos:
// @return []string:
func GetSelectorFromPOs(pos any) ([]string, error) {
	rt := reflect.TypeOf(pos)
	structType, err := greflect.GetElemStructType(rt)
	if err != nil {
		return nil, err
	}
	return getSelectorFromStructType(structType)
}

func getSelectorFromStructType(structType reflect.Type) ([]string, error) {
	value, ok := type2Selector.Load(structType)
	if ok {
		return value.([]string), nil
	}
	return getSelectorFromStructTypeSlow(structType)
}

// getSelectorFromStructTypeSlow
// @Description: 不用缓存，读取 structType 的 gorm tag，构造 []string
// @param structType:
// @return []string:
func getSelectorFromStructTypeSlow(structType reflect.Type) ([]string, error) {
	if structType.Kind() != reflect.Struct {
		return nil, gerror.GetSelectorFromNonStructErr(structType)
	}
	var selector []string
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		gormTag := strings.TrimSpace(structField.Tag.Get("gorm"))
		tagKVs, err := getKVsFromTag(gormTag)
		if err != nil {
			return nil, err
		}
		columnName := tagKVs["column"]
		selector = append(selector, columnName)
	}
	type2Selector.Store(structType, selector)
	return selector, nil
}

// getKVsFromTag
// @Description: 从 `k1:v1;k2:v2;...` 格式的 tag 中读取 kv map
// @param tag:
// @return map[string]string:
func getKVsFromTag(tag string) (map[string]string, error) {
	kvs := strings.Split(tag, ";")
	kvMap := make(map[string]string, len(kvs))
	for _, s := range kvs {
		kv := strings.Split(s, ":")
		if len(kv) != 2 {
			return nil, gerror.GormTagShouldBeKVsErr(tag)
		}
		kvMap[kv[0]] = kv[1]
	}
	return kvMap, nil
}
