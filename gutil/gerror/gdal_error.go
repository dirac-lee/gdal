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
