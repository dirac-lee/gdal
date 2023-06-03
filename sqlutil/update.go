package sqlutil

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

// BuildSQLUpdate
//
// @Description: 将 Update model struct 编译为 sql 语句中 update 的字典
//
// @param update: Update model
//
// @return m:
//
// @return err:
//
// @example：
//
//	    // model 表
//	    type TableAbc struct {
//	        ID   int64  `gorm:"column:id"`
//	        Name string `gorm:"column:name"`
//	        Age  int    `gorm:"column:p_age"`
//	    }
//
//	    func (Campaign) TableName() string {
//	    	return "table_abc"
//	    }
//
//	    // 需要更新的字段
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
//	    // 下面即 sql： update table_abc set name="byte-er" where id = 1
//		attrs, err := BuildSQLUpdate(attrs)
//	    if err != nil{
//	        // do something
//	    }
//	    if err := db.Model(TableAbc{}).Where("id = ?", id).Updates(attrs).Error; err != nil {
//	        logs.Error("update table abc failed: %s", err)
//	    }
func BuildSQLUpdate(update any) (map[string]any, error) {
	rv, rt, err := GetElemValueTypeOfPtr(reflect.ValueOf(update))
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

// 遍历 field，将非 nil 的值拼到 map 中
func fillSQLUpdateFieldMap(rv reflect.Value, st *sqlType) (map[string]any, error) {
	m := make(map[string]any)
	for _, name := range st.Names {
		column := st.ColumnsMap[name] // 前置函数已经检查过，一定存在
		data := rv.FieldByName(column.Name)
		// 字段的值是 nil 直接忽略 不做处理
		if data.Kind() == reflect.Ptr && data.IsNil() {
			continue
		}
		if data.Kind() == reflect.Ptr {
			data = data.Elem()
		}
		if column.Expr != "" {
			updater := updaterMap[column.Expr] // 前置函数已经检查过，一定存在
			if updaterResult := updater(column.Field, data.Interface()); updaterResult.SQL != "" {
				m[column.Field] = updaterResult
			}
		} else {
			m[column.Field] = data.Interface()
		}
	}

	return m, nil
}

// SQLUpdater update语句生成器
type SQLUpdater func(field string, data any) clause.Expr

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

func isMergeJSONStruct(v any) bool {
	vt := reflect.TypeOf(v)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	return vt.Kind() == reflect.Struct
}

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
