package gerror

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	GDALErr = errors.New("[GDAL] error")
)

func GDALErrorf(format string, a ...any) error {
	return fmt.Errorf("%w: %s", GDALErr, fmt.Sprintf(format, a...))
}

func NonStructBasedModelErr(rt reflect.Type) error {
	return GDALErrorf("non-struct-based model (%v) is not supported", rt)
}

func InvalidReflectValueErr(rv reflect.Value) error {
	return GDALErrorf("invalid reflect.Value (%v)", rv)
}

func GormTagShouldBeKVsErr(tag string) error {
	return GDALErr("gorm tag should be kvs, bug got: %v", tag)
}

func GetSelectorFromNonStructErr(rt reflect.Type) error {
	return GDALErr("could not get selector from non-struct type: %v", rt)
}
