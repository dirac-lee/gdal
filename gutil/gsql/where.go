package gsql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dirac-lee/gdal/gutil/greflect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func BuildSQLWhereV2(where any) (clause.Where, error) {
	rv, rt, err := greflect.GetElemValueTypeOfPtr(reflect.ValueOf(where))
	if err != nil {
		return clause.Where{}, err
	}
	return buildSQLWhereV2(rv, rt)
}

func buildSQLWhereV2(rv reflect.Value, rt reflect.Type) (clause.Where, error) {
	var exprs []clause.Expression
	firstExprs, err := buildSQLAndExprsV2(rv, rt)
	if err != nil {
		return clause.Where{}, err
	}

	orClauseList := getOrClauseList(rv, rt)
	if len(orClauseList) == 0 {
		return clause.Where{Exprs: firstExprs}, nil
	}
	if len(firstExprs) > 0 {
		exprs = append(exprs, clause.And(firstExprs...))
	}
	for _, orClause := range orClauseList { // multiple fields with tag $or
		if !orClause.IsValid() {
			continue
		}
		orExprs, err := buildSQLOrExprsV2(orClause)
		if err != nil {
			return clause.Where{}, err
		}
		if len(orExprs) > 0 {
			orClauseExpr := clause.Or(orExprs...) // use OR to combine elem of slice field with tag $or
			exprs = append(exprs, orClauseExpr)   // use AND to combine fields with tag $or
		}
	}
	return clause.Where{Exprs: exprs}, nil
}

func buildSQLOrExprsV2(rv reflect.Value) ([]clause.Expression, error) {
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice { // 每个 $or 都必须是 slice
		return nil, errors.New("or clauses must be slice or array")
	}

	exprs := make([]clause.Expression, 0, rv.Len())
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
		subWhere, err := buildSQLWhereV2(erv, ert) // embedding
		if err != nil {
			return nil, err
		}
		if len(subWhere.Exprs) == 0 {
			continue
		}
		exprs = append(exprs, clause.And(subWhere.Exprs...))
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
		expr := fmt.Sprintf("`%v` < ?", column)
		return gorm.Expr(expr, data), nil
	},
	"<=": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` <= ?", column)
		return gorm.Expr(expr, data), nil
	},
	"=": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` = ?", column)
		return gorm.Expr(expr, data), nil
	},
	"": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` = ?", column)
		return gorm.Expr(expr, data), nil
	},
	"!=": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` != ?", column)
		return gorm.Expr(expr, data), nil
	},
	">": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` > ?", column)
		return gorm.Expr(expr, data), nil
	},
	">=": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` >= ?", column)
		return gorm.Expr(expr, data), nil
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
		expr := fmt.Sprintf("`%v` LIKE ?", column)
		return gorm.Expr(expr, "%"+data.(string)+"%"), nil
	},
	"left like": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` LIKE ?", column)
		return gorm.Expr(expr, "%"+data.(string)), nil
	},
	"right like": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` LIKE ?", column)
		return gorm.Expr(expr, data.(string)+"%"), nil
	},
	"like": func(column string, data any) (clause.Expression, error) {
		expr := fmt.Sprintf("`%v` LIKE ?", column)
		return gorm.Expr(expr, data.(string)), nil
	},
	"json_contains": func(column string, data any) (clause.Expression, error) {
		return JSONContains(column, data), nil
	},
	"json_contains any": func(column string, data any) (clause.Expression, error) {
		exprs, err := JSONContainsExprs(column, data)
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
