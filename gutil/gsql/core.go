package gsql

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var cacheMap sync.Map

type sqlType struct {
	Names      []string
	ColumnsMap map[string]*sqlColumn
}

type sqlColumn struct {
	Index       int          // field 序号
	Name        string       // field name
	Field       string       // tag sql_field
	Operator    string       // tag sql_operator
	Expr        string       // tag sql_expr
	IsAnonymous bool         // field 是否是匿名字段
	Kind        reflect.Kind // field Kind
}

func parseType(t reflect.Type) (*sqlType, error) {
	sqlType := loadTypeFromCache(t)
	if sqlType != nil {
		return sqlType, nil
	}
	parsedType, err := parseTypeSlow(t)
	if err != nil {
		return nil, err
	}
	cacheMap.Store(t, parsedType)
	return parsedType, nil
}

func loadTypeFromCache(t reflect.Type) *sqlType {
	v, ok := cacheMap.Load(t)
	if ok {
		sqlType := v.(*sqlType)
		return sqlType
	}
	return nil
}

func parseTypeSlow(structType reflect.Type) (_ *sqlType, err error) {
	result, err := parseTypeRev(structType, nil, false)
	if err != nil {
		return nil, err
	}
	for _, name := range result.Names {
		column := result.ColumnsMap[name]
		if column == nil {
			return nil, fmt.Errorf("field(%s) column not found", name)
		}
		if column.Expr != "" {
			if _, ok := updaterMap[column.Expr]; !ok {
				return nil, fmt.Errorf("field(%s) operator(%s) invalid", column.Name, column.Operator)
			}
		}
	}
	return result, nil
}

func parseTypeRev(structType reflect.Type, sType *sqlType, isField bool) (_ *sqlType, err error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if isField && (structType.Kind() != reflect.Struct && structType.Kind() != reflect.Slice) {
		return nil, fmt.Errorf("opt kind must be struct, but got %s", structType.Kind())
	}

	if sType == nil {
		sType = &sqlType{
			ColumnsMap: map[string]*sqlColumn{},
		}
	}
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		sqlField := strings.TrimSpace(structField.Tag.Get("sql_field"))
		sqlOperator := strings.TrimSpace(structField.Tag.Get("sql_operator"))
		sqlExpr := strings.TrimSpace(structField.Tag.Get("sql_expr"))
		// 忽略 tag
		if sqlField == "-" || (sqlField == "" && sqlExpr == "$or") {
			continue
		}
		if sqlExpr == "$or" {
			return nil, fmt.Errorf("struct field(%s) with mix of sql_field(%v) and expr(%s) invalid", structField.Name, sqlField, sqlExpr)
		}
		if err := checkField(structField, sqlField); err != nil {
			return nil, err
		}
		if structField.Anonymous {
			// 内嵌结构体，递归处理
			t := structField.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			childSType, err := parseTypeRev(t, sType, true)
			if err != nil {
				return nil, err
			}
			for kk, vv := range childSType.ColumnsMap {
				if _, ok := sType.ColumnsMap[kk]; !ok {
					sType.ColumnsMap[kk] = vv
					sType.Names = append(sType.Names, kk)
				}
			}
		} else {
			if sqlOperator != "" {
				if err := checkOperator(structField, sqlOperator); err != nil {
					return nil, err
				}
			}
			if sqlExpr != "" {
				if _, ok := updaterMap[sqlExpr]; !ok {
					return nil, fmt.Errorf("field(%s) expr(%s) invalid", structField.Name, sqlExpr)
				}
			}
			column := &sqlColumn{
				Index:       i,
				Name:        structField.Name,
				Field:       sqlField,
				Operator:    sqlOperator,
				Expr:        sqlExpr,
				IsAnonymous: structField.Anonymous,
				Kind:        structField.Type.Kind(),
			}
			// 已经有这个 name 的 field 这说明需要覆盖
			if _, ok := sType.ColumnsMap[structField.Name]; ok {
				reOrderNames(sType, structField.Name)
			} else {
				sType.Names = append(sType.Names, structField.Name)
			}
			sType.ColumnsMap[structField.Name] = column
		}
	}
	return sType, nil
}

func reOrderNames(sType *sqlType, name string) {
	for idx, n := range sType.Names {
		if n == name {
			copy(sType.Names[idx:], sType.Names[idx+1:])
			sType.Names[len(sType.Names)-1] = name
			break
		}
	}
}

func checkField(structField reflect.StructField, sqlField string) error {
	if structField.Anonymous && sqlField != "" {
		return fmt.Errorf("field %s is anonymous that can not have sql_field tag", structField.Name)
	}

	if !structField.Anonymous {
		if structField.Type.Kind() != reflect.Ptr && structField.Type.Kind() != reflect.Slice {
			return fmt.Errorf("struct field(%s) must be pointer, but got %s", structField.Name, structField.Type.Kind())
		}
		if sqlField == "" {
			return fmt.Errorf("struct field(%s) need sql_field tag", structField.Name)
		}
	}

	return nil
}

func checkOperator(field reflect.StructField, sqlOperator string) error {
	if _, ok := whereMap[sqlOperator]; !ok {
		return fmt.Errorf("field(%s) operator(%s) invalid", field.Name, sqlOperator)
	}

	if field.Type.Kind() != reflect.Ptr {
		// in 操作，是数组，不是指针
		if field.Type.Kind() == reflect.Slice && isOperatorSupportArray(sqlOperator) {
			// continue
		} else if field.Type.Kind() == reflect.Slice && field.Name == "Select" {
			// continue
		} else {
			return fmt.Errorf("struct field(%s) must be pointer, but got %s", field.Name, field.Type.Kind())
		}
	}
	return nil
}

func isOperatorSupportArray(s string) bool {
	return s == "in" || s == "not in" || s == "json_contains" || s == "json_contains any" || s == "json_contains all"
}
