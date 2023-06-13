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

// MakeQueryConfig convert options to QueryConfig
func MakeQueryConfig(options []QueryOption) *QueryConfig {
	opt := new(QueryConfig)
	for _, v := range options {
		if v != nil {
			v(opt)
		}
	}
	return opt
}

// WithLimit assign limit
func WithLimit(limit int) QueryOption {
	return func(v *QueryConfig) {
		v.Limit = &limit
	}
}

// WithOffset assign offset
func WithOffset(offset int) QueryOption {
	return func(v *QueryConfig) {
		v.Offset = &offset
	}
}

// WithOrder assign order
func WithOrder(order string) QueryOption {
	return func(v *QueryConfig) {
		v.Order = &order
	}
}

// WithSelects assign selected columns
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func WithSelects(selects []string) QueryOption {
	return func(v *QueryConfig) {
		v.Selects = selects
	}
}

// WithMaster read master
//
// 💡 HINT:
//
// ⚠️  WARNING:
//
// 🚀 example:
func WithMaster() QueryOption {
	return func(v *QueryConfig) {
		v.readMaster = true
	}
}
