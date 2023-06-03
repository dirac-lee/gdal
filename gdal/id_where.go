package gdal

type idWhere struct {
	ID   *int64  `sql_field:"id" sql_operator:"="`
	IDIn []int64 `sql_field:"id" sql_operator:"in"`
}
