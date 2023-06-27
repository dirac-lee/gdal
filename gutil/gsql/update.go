package gsql

import (
	"fmt"
	"github.com/dirac-lee/gdal/gutil/greflect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
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
type SQLUpdater func(column string, data any) clause.Expr

// updaterMap support `+`, `-` and `merge_json` so far
var updaterMap = map[string]SQLUpdater{
	"+": func(column string, data any) clause.Expr {
		return gorm.Expr(column+" + ?", data)
	},
	"-": func(column string, data any) clause.Expr {
		return gorm.Expr(column+" - ?", data)
	},
	"json_set": func(column string, data any) clause.Expr {
		return JSONSetExpr(column, data)
	},
}

func JSONSetExpr(column string, data any) clause.Expr {
	reflectValue := reflect.ValueOf(data)
	reflectType := reflectValue.Type()
	sqlKey := make([]string, 0)
	sqlVal := make([]interface{}, 0)
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		value := reflectValue.Field(i)

		if value.Kind() == reflect.Ptr && value.IsNil() {
			continue
		}
		if value.Kind() == reflect.Slice && (value.IsNil() || value.Len() == 0) {
			continue
		}
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		sqlVal = append(sqlVal, value.Interface())

		// get json tag as sql key if exists, otherwise, use field name
		jsonTag := field.Tag.Get("json")
		if len(jsonTag) > 0 {
			sqlKey = append(sqlKey, jsonTag)
		} else {
			sqlKey = append(sqlKey, field.Name)
		}
	}

	if len(sqlVal) <= 0 {
		return clause.Expr{}
	}
	var sqlStr string
	for _, key := range sqlKey {
		sqlStr += fmt.Sprintf(", '$.%s', ?", key)
	}
	return clause.Expr{
		SQL:  fmt.Sprintf("JSON_SET(`%s` %s)", column, sqlStr),
		Vars: sqlVal,
	}
}
