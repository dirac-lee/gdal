package gsql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/dirac-lee/gdal/gutil/greflect"
)

type FieldExpr string // sql: field_1 > field_2, name field_2 need use FieldExpr type

// BuildSQLWhere build Where model struct into query & args in SQL
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
//
//	// model table
//	type TableAbc struct {
//		ID   int64  `gorm:"column:id"`
//		Name string `gorm:"column:name"`
//		Age  int    `gorm:"column:p_age"`
//	}
//
//	func (TableAbc) TableName() string {
//		return "table_abc"
//	}
//
//	// fields to be updated
//	type TableAbcWhere struct {
//		Name *string `sql_field:"name" sql_operator:"like"`
//		Age  *int    `sql_field:"p_age"`
//	}
//
//	func example() {
//		var name = "name%"
//		var age = 20
//		attrs := TableAbcWhere{
//			Name: &name,
//			Age:  &age,
//		}
//
//		query, args, err := BuildSQLWhere(attrs)
//		if err != nil {
//			// handle error
//		}
//
//		// SQL： update table_abc set name="byte-er" where id = 1
//		if err := db.Find(&pos).Where(query, args...).Error; err != nil {
//			logs.Error("fins table abc failed: %s", err)
//		}
//	}
func BuildSQLWhere(where any) (query string, args []any, err error) {
	rv, rt, err := greflect.GetElemValueTypeOfPtr(reflect.ValueOf(where))
	if err != nil {
		return "", nil, err
	}
	return buildSQLWhere(rv, rt)
}

func buildSQLWhere(rv reflect.Value, rt reflect.Type) (query string, args []any, err error) {
	andPrefix, andArgs, err := buildSQLWhereWithAndOption(rv, rt)
	if err != nil {
		return "", nil, err
	}

	// 使用 and 拼接 queries 条件
	var queries []string
	if len(andPrefix) > 0 {
		queries = append(queries, andPrefix)
		args = append(args, andArgs...)
	}

	orClauseList := getOrClauseList(rv, rt)
	for _, orClause := range orClauseList {
		if !orClause.IsValid() {
			continue
		}

		// connect clauses below by `or`
		orSuffix, orArgs, err := buildSQLWhereWithOrOptions(orClause)
		if err != nil {
			return "", nil, err
		}

		if len(orSuffix) > 0 { // use ( ) to embrace the conditions connected with `or` if exists
			queries = append(queries, fmt.Sprintf("(%v)", orSuffix))
		}
		args = append(args, orArgs...)
	}

	query = strings.Join(queries, " and ")

	return query, args, err
}

var orCache sync.Map

func getOrClauseList(rv reflect.Value, rt reflect.Type) (orClauses []reflect.Value) {
	var orIndices []int
	value, cached := orCache.Load(rt)
	if cached {
		orIndices = value.([]int)
		for _, index := range orIndices {
			fieldValue := rv.Field(index)
			orClauses = append(orClauses, fieldValue)
		}
		return orClauses
	}

	for i := 0; i < rt.NumField(); i++ {
		fieldValue := rv.Field(i)
		structField := rt.Field(i)
		sqlExpr := strings.TrimSpace(structField.Tag.Get("sql_expr"))
		if sqlExpr == "$or" {
			orIndices = append(orIndices, i)
			orClauses = append(orClauses, fieldValue)
		}
	}
	orCache.Store(rt, orIndices)
	return orClauses
}

func buildSQLWhereWithOrOptions(rv reflect.Value) (query string, args []any, err error) {
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return "", nil, errors.New("or clauses must be slice or array")
	}

	subQueries := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ { // 遍历 orOpts 数组
		erv := rv.Index(i)
		if !erv.IsValid() {
			continue
		}
		if erv.Kind() == reflect.Ptr {
			erv = erv.Elem()
		}
		if erv.Kind() != reflect.Struct {
			continue
		}
		ert := erv.Type()
		subQuery, subArgs, err := buildSQLWhere(erv, ert) // 支持嵌套
		if err != nil {
			return "", nil, err
		}
		if len(subQuery) == 0 {
			continue
		}
		if len(subArgs) > 1 {
			subQuery = fmt.Sprintf("(%v)", subQuery)
		}
		subQueries = append(subQueries, subQuery)
		args = append(args, subArgs...)
	}

	query = strings.Join(subQueries, " or ")

	return query, args, err
}

func buildSQLWhereWithAndOption(rv reflect.Value, rt reflect.Type) (query string, args []any, err error) {
	// 遍历 field，将非 nil 的值拼到 map 中
	query, args, err = fillSQLWhereCondition(rv, rt)
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

// fillSQLWhereCondition walk through all the fields in `rv`, parsed to single where conditions, then join them with `AND`.
//
// 💡 HINT:
//
// ⚠️  WARNING: empty slice []T{} is treated as zero value.
//
// 🚀 example:
//
//
func fillSQLWhereCondition(rv reflect.Value, rt reflect.Type) (query string, args []any, err error) {
	args = []any{}
	qq := new(strings.Builder)
	isFirst := true

	sqlType, err := parseType(rt)
	if err != nil {
		return "", nil, err
	}

	for _, name := range sqlType.Names {
		column := sqlType.ColumnsMap[name] // 前置步骤检查过，一定存在
		data := rv.FieldByName(column.Name)
		// 字段的值是 nil 直接忽略 不做处理
		if data.Kind() == reflect.Ptr && data.IsNil() {
			continue
		}
		// slice 长度为 0 也直接忽略 不做处理
		if data.Kind() == reflect.Slice && (data.IsNil() || data.Len() == 0) {
			continue
		}
		if data.Kind() == reflect.Ptr {
			data = data.Elem()
		}
		inter := data.Interface()
		op, err := GetOperatorMap(column.Operator, inter)
		if err != nil {
			return "", nil, err
		}
		operator, placeholder, arg := op(inter)
		if isFirst {
			isFirst = false
		} else {
			qq.WriteString(" and ")
		}

		// 支持跨表查询时，表中有相同字段的情况, example: `sql_field:"others.id"`
		// 如果 sqlField 中存在 '.', 则 split by '.', 并在字段前后加上 '`', 重新用 '.' 拼接起来
		sqlField := strings.Replace(column.Field, ".", "`.`", 1)

		var fieldExpr string
		if arg != nil {
			switch d := arg.(type) {
			case FieldExpr:
				fieldExpr = string(d)
			case *FieldExpr:
				if d != nil {
					fieldExpr = string(*d)
				}
			}
		}

		qq.WriteString("`")
		qq.WriteString(sqlField)
		qq.WriteString("` ")
		qq.WriteString(operator)
		qq.WriteString(" ")
		if fieldExpr != "" {
			qq.WriteString("`")
			qq.WriteString(fieldExpr)
			qq.WriteString("`")
		} else {
			qq.WriteString(placeholder)
			if arg != nil {
				args = append(args, arg)
			}
		}
	}

	return qq.String(), args, nil
}

type (
	Operator       func(data any) (operator string, placeholder string, arg any)
	OperatorFilter func(data any) bool
)

var operatorMap = map[string]struct {
	Operator       Operator
	OperatorFilter OperatorFilter
}{
	"<": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "<", "?", data
		},
	},
	"<=": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "<=", "?", data
		},
	},
	"=": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "=", "?", data
		},
	},
	"": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "=", "?", data
		},
	},
	"!=": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "!=", "?", data
		},
	},
	">": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return ">", "?", data
		},
	},
	">=": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return ">=", "?", data
		},
	},
	"null": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			switch v := data.(type) {
			case bool:
				if v {
					return "is null", "", nil
				} else {
					return "is not null", "", nil
				}
			}
			return "null", "", data
		},
	},
	"in": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "in", "(?)", data
		},
	},
	"not in": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "not in", "(?)", data
		},
	},
	"full like": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "like", "?", "%" + data.(string) + "%"
		},
		OperatorFilter: isStringEmpty,
	},
	"left like": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "like", "?", "%" + data.(string)
		},
		OperatorFilter: isStringEmpty,
	},
	"right like": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "like", "?", data.(string) + "%"
		},
		OperatorFilter: isStringEmpty,
	},
	"like": {
		Operator: func(data any) (operator string, placeholder string, arg any) {
			return "like", "?", data.(string)
		},
		OperatorFilter: isStringEmpty,
	},
}

func isStringEmpty(data any) bool {
	switch r := data.(type) {
	case string:
		return r != ""
	}
	return true
}

func GetOperatorMap(operatorKey string, data any) (Operator, error) {
	operator, ok := operatorMap[operatorKey]
	if !ok {
		return nil, fmt.Errorf("operator %q not found", operatorKey)
	}
	if operator.OperatorFilter == nil {
		return operator.Operator, nil
	}
	if operator.OperatorFilter(data) {
		return operator.Operator, nil
	}
	return nil, fmt.Errorf("operator %q not found", operatorKey)
}
