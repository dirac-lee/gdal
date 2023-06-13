package gsql

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/dirac-lee/gdal/gutil/greflect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BuildSQLUpdate build Update struct into sql update map
//
// 💡 HINT:
//
// ⚠️  WARNING: fields of update must be pointers
//
// 🚀 example:
//
//	    // model of table_abc
//	    type TableAbc struct {
//	        ID   int64  `gorm:"column:id"`
//	        Name string `gorm:"column:name"`
//	        Age  int    `gorm:"column:p_age"`
//	    }
//
//	    func (TableAbc) TableName() string {
//	    	return "table_abc"
//	    }
//
//	    // fields need to update.
//	    type TableAbcUpdate struct {
//	        Name *string `sql_field:"name"`
//	        Age  *int    `sql_field:"p_age"`
//	    }
//
//	    var name = "byte-er"
//	    attrs = TableAbcUpdate{
//	        Name: &name
//	    }
//
//	    // SQL: update table_abc set name="byte-er" where id = 1
//		attrs, err := BuildSQLUpdate(attrs)
//	    if err != nil{
//	        // do something
//	    }
//	    if err := db.Model(TableAbc{}).Where("id = ?", id).Updates(attrs).Error; err != nil {
//	        logs.Error("update table abc failed: %s", err)
//	    }
func BuildSQLUpdate(update any) (map[string]any, error) {
	rv, rt, err := greflect.GetElemValueTypeOfPtr(reflect.ValueOf(update))
	if err != nil {
		return nil, err
	}
	// 针对类型的检查 解析的时候有做
	sqlType, err := parseType(rt)
	if err != nil {
		return nil, err
	}

	// 遍历 field，将非 nil 的值拼到 map 中
	m, err := fillSQLUpdateFieldMap(rv, sqlType)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// fillSQLUpdateFieldMap walk through all the fields in `rv`，insert non-zero fields into map.
//
// 💡 HINT:
//
// ⚠️  WARNING: empty slice []T{} is treated as zero value.
//
// 🚀 example:
//
//
func fillSQLUpdateFieldMap(rv reflect.Value, st *sqlType) (map[string]any, error) {
	m := make(map[string]any)
	for _, name := range st.Names {
		column := st.ColumnsMap[name] // must be found, guaranteed by previous operations
		data := rv.FieldByName(column.Name)
		// skip nil
		if data.Kind() == reflect.Ptr && data.IsNil() {
			continue
		}
		// only supports one-level pointers
		if data.Kind() == reflect.Ptr {
			data = data.Elem()
		}
		if column.Expr != "" {
			updater := updaterMap[column.Expr] // must be found, guaranteed by previous operations
			if updaterResult := updater(column.Field, data.Interface()); updaterResult.SQL != "" {
				m[column.Field] = updaterResult
			}
		} else {
			m[column.Field] = data.Interface()
		}
	}

	return m, nil
}

// SQLUpdater update SQL generator
type SQLUpdater func(field string, data any) clause.Expr

// updaterMap support `+`, `-` and `merge_json` so far
var updaterMap = map[string]SQLUpdater{
	"+": func(field string, data any) clause.Expr {
		return gorm.Expr(field+" + ?", data)
	},
	"-": func(field string, data any) clause.Expr {
		return gorm.Expr(field+" - ?", data)
	},
	"merge_json": func(field string, data any) clause.Expr {
		var bs []byte
		if isMergeJSONStruct(data) {
			dataMap, _ := mergeJSONStructToJSONMap(data)
			bs, _ = json.Marshal(dataMap)
		} else {
			bs, _ = json.Marshal(data)
		}
		s := string(bs)
		if s == "" {
			return clause.Expr{}
		}
		return gorm.Expr("CASE WHEN (`"+field+"` IS NULL OR `"+field+"` = '') THEN CAST(? AS JSON) ELSE JSON_MERGE_PATCH(`"+field+"`, CAST(? AS JSON)) END", s, s)
	},
}

// isMergeJSONStruct whether `v` can be a struct or a pointer to struct
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
//
//
func isMergeJSONStruct(v any) bool {
	vt := reflect.TypeOf(v)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	return vt.Kind() == reflect.Struct
}

// mergeJSONStructToJSONMap convert struct to map by tag `json`
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
//
//
func mergeJSONStructToJSONMap(v any) (map[string]any, error) {
	vt := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)

	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
		vv = vv.Elem()
	}
	if vt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("update(JSON_MERGE_PATCH) need struct type")
	}

	m := map[string]any{}
	for i := 0; i < vt.NumField(); i++ {
		vtField := vt.Field(i)
		vvField := vv.Field(i)

		if !vvField.IsValid() {
			continue
		}

		jsonField := vtField.Tag.Get("json")
		if jsonField == "" {
			continue
		}

		// ptr
		if vtField.Type.Kind() == reflect.Ptr {
			if vvField.IsNil() {
				continue
			}
			m[jsonField] = vvField.Elem().Interface()
		} else {
			m[jsonField] = vvField.Interface()
		}
	}

	return m, nil
}
