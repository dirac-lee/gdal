package gdal

type QueryConfig struct {
	readMaster bool
	debug      bool

	// export field
	Limit   *int
	Offset  *int
	Order   *string
	Selects []string
}

type QueryOption func(v *QueryConfig)

// MakeQueryConfig
// @Description: 将 options 转化为 QueryConfig
// @param options:
// @return *QueryConfig:
func MakeQueryConfig(options []QueryOption) *QueryConfig {
	opt := new(QueryConfig)
	for _, v := range options {
		if v != nil {
			v(opt)
		}
	}
	return opt
}

// WithLimit
// @Description: 指定限制条数
// @param limit: 限制条数
// @return QueryOption:
func WithLimit(limit int) QueryOption {
	return func(v *QueryConfig) {
		v.Limit = &limit
	}
}

// WithOffset
// @Description: 指定查询偏移
// @param offset: 查询偏移
// @return QueryOption:
func WithOffset(offset int) QueryOption {
	return func(v *QueryConfig) {
		v.Offset = &offset
	}
}

// WithOrder
// @Description: 指定查询顺序
// @param order: 查询顺序
// @return QueryOption:
func WithOrder(order string) QueryOption {
	return func(v *QueryConfig) {
		v.Order = &order
	}
}

// WithSelects
// @Description: 指定查询字段
// @param selects: 查询字段
// @return QueryOption:
func WithSelects(selects []string) QueryOption {
	return func(v *QueryConfig) {
		v.Selects = selects
	}
}

// WithMaster
// @Description: 查主库
// @return QueryOption:
func WithMaster() QueryOption {
	return func(v *QueryConfig) {
		v.readMaster = true
	}
}

// WithDebug
// @Description: Debug 模式
// @return QueryOption:
func WithDebug() QueryOption {
	return func(v *QueryConfig) {
		v.debug = true
	}
}
