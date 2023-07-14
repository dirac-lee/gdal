package gsql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/dirac-lee/gdal/gutil/greflect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BuildSQLWhereExpr build Where model struct into query & args in SQL
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
//		gwhere, err := BuildSQLWhereExpr(attrs)
//		if err != nil {
//			// handle error
//		}
//
//		// SQL： update table_abc set name="byte-er" where id = 1
//		if err := db.Find(&pos).Where(gwhere).Error; err != nil {
//			logs.Error("fins table abc failed: %s", err)
//		}
//	}
func BuildSQLWhereExpr(where any) (clause.Expression, error) {
	rv, rt, err := greflect.GetElemValueTypeOfPtr(reflect.ValueOf(where))
	if err != nil {
		return nil, err
	}
	return buildSQLWhereV2(rv, rt)
}

func buildSQLWhereV2(rv reflect.Value, rt reflect.Type) (clause.Expression, error) {
	var exprs []clause.Expression
	firstExprs, err := buildSQLAndExprsV2(rv, rt)
	if err != nil {
		return nil, err
	}
	if len(firstExprs) > 0 {
		exprs = append(exprs, clause.And(firstExprs...))
	}

	orClauseList := getOrClauseList(rv, rt)
	if len(orClauseList) == 0 {
		return clause.And(firstExprs...), nil
	}
	for _, orClause := range orClauseList { // multiple fields with tag $or
		if !orClause.IsValid() {
			continue
		}
		orExprs, err := buildSQLOrExprsV2(orClause)
		if err != nil {
			return nil, err
		}
		if len(orExprs) > 0 {
			orClauseExpr := clause.Or(orExprs...) // use OR to combine elem of slice field with tag $or
			exprs = append(exprs, orClauseExpr)   // use AND to combine fields with tag $or
		}
	}
	if len(exprs) == 0 {
		return nil, nil
	}
	return clause.And(exprs...), nil
}

func buildSQLOrExprsV2(rv reflect.Value) ([]clause.Expression, error) {
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice { // 每个 $or 都必须是 slice
		return nil, errors.New("or clauses must be slice or array")
	}

	var exprs []clause.Expression
	for i := 0; i < rv.Len(); i++ { // range slice field with tag $or
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
		subExpr, err := buildSQLWhereV2(erv, ert) // embedding
		if err != nil {
			return nil, err
		}
		if subExpr == nil {
			continue
		}
		exprs = append(exprs, clause.And(subExpr))
	}

	return exprs, nil
}

func buildSQLAndExprsV2(rv reflect.Value, rt reflect.Type) ([]clause.Expression, error) {
	var exprs []clause.Expression

	sqlType, err := parseType(rt)
	if err != nil {
		return nil, err
	}

	for _, name := range sqlType.Names {
		column := sqlType.ColumnsMap[name]
		field := rv.FieldByName(column.Name)
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}
		if field.Kind() == reflect.Slice && (field.IsNil() || field.Len() == 0) {
			continue
		}
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}
		data := field.Interface()

		builder, err := GetWhereExpr(column.Operator)
		if err != nil {
			return nil, err
		}
		sqlField := strings.Replace(column.Field, ".", "`.`", 1)
		expr, err := builder(sqlField, data)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	return exprs, nil
}

func GetWhereExpr(operator string) (SQLWhereExprBuilder, error) {
	expr, ok := whereMap[operator]
	if !ok {
		return nil, fmt.Errorf("unsupported operator %s", operator)
	}
	return expr, nil
}

// SQLWhereExprBuilder where SQL generator
type SQLWhereExprBuilder func(column string, data any) (clause.Expression, error)

var whereMap = map[string]SQLWhereExprBuilder{
	"<": func(column string, data any) (clause.Expression, error) {
		return clause.Lt{Column: column, Value: data}, nil
	},
	"<=": func(column string, data any) (clause.Expression, error) {
		return clause.Lte{Column: column, Value: data}, nil
	},
	"=": func(column string, data any) (clause.Expression, error) {
		return clause.Eq{Column: column, Value: data}, nil
	},
	"": func(column string, data any) (clause.Expression, error) {
		return clause.Eq{Column: column, Value: data}, nil
	},
	"!=": func(column string, data any) (clause.Expression, error) {
		return clause.Neq{Column: column, Value: data}, nil
	},
	"<>": func(column string, data any) (clause.Expression, error) {
		return clause.Neq{Column: column, Value: data}, nil
	},
	">": func(column string, data any) (clause.Expression, error) {
		return clause.Gt{Column: column, Value: data}, nil
	},
	">=": func(column string, data any) (clause.Expression, error) {
		return clause.Gte{Column: column, Value: data}, nil
	},
	"null": func(column string, data any) (clause.Expression, error) {
		v, isBool := data.(bool)
		if !isBool {
			return clause.Expr{}, errors.New("field with tag `null` must be bool")
		}
		var not string
		if !v {
			not = "NOT "
		}
		expr := fmt.Sprintf("`%v` IS %vNULL", column, not)
		return gorm.Expr(expr), nil
	},
	"in": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` IN (?)", column)
		return gorm.Expr(expr, data), nil
	},
	"not in": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` NOT IN (?)", column)
		return gorm.Expr(expr, data), nil
	},
	"full like": func(column string, data any) (clause.Expression, error) {
		v, isStr := data.(string)
		if !isStr {
			return clause.Expr{}, errors.New("field with tag `full like` must be string")
		}
		return clause.Like{Column: column, Value: "%" + v + "%"}, nil
	},
	"left like": func(column string, data any) (clause.Expression, error) {
		v, isStr := data.(string)
		if !isStr {
			return clause.Expr{}, errors.New("field with tag `left like` must be string")
		}
		return clause.Like{Column: column, Value: "%" + v}, nil
	},
	"right like": func(column string, data any) (clause.Expression, error) {
		v, isStr := data.(string)
		if !isStr {
			return clause.Expr{}, errors.New("field with tag `right like` must be string")
		}
		return clause.Like{Column: column, Value: v + "%"}, nil
	},
	"like": func(column string, data any) (clause.Expression, error) {
		v, isStr := data.(string)
		if !isStr {
			return clause.Expr{}, errors.New("field with tag `like` must be string")
		}
		return clause.Like{Column: column, Value: v}, nil
	},
	"json_contains": func(column string, data any) (clause.Expression, error) {
		return JSONContains(column, data), nil
	},
	"json_contains any": func(column string, data any) (clause.Expression, error) {
		exprs, err := JSONContainsExprs(column, data)
		if len(exprs) == 1 {
			return exprs[0], nil
		}
		return clause.Or(exprs...), err
	},
	"json_contains all": func(column string, data any) (clause.Expression, error) {
		exprs, err := JSONContainsExprs(column, data)
		return clause.And(exprs...), err
	},
}

func JSONContains(column string, data any) clause.Expression {
	expr := fmt.Sprintf("JSON_CONTAINS(%s, ?)", column)
	return gorm.Expr(expr, data)
}

func JSONContainsExprs(column string, data any) ([]clause.Expression, error) {
	rv := reflect.ValueOf(data)
	rt := rv.Type()
	if rt.Kind() != reflect.Slice && rt.Kind() != reflect.Array {
		return nil, errors.New("field with tag `json_contains any` must be slice or array")
	}
	var exprs []clause.Expression
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		expr := JSONContains(column, elem.Interface())
		exprs = append(exprs, expr)
	}
	return exprs, nil
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
