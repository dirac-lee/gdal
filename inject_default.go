package gdal

// InjectDefaulter inject default for *Where
//
// ⚠️  WARNING: please implement InjectDefaulter for *Where in spite of Where
type InjectDefaulter interface {
	InjectDefault()
}

// injectDefaultIfHas inject default for *Where if it implements InjectDefaulter
//
// 💡 HINT:
//
// ⚠️  WARNING: please implement InjectDefaulter for *Where in spite of Where
//
// 🚀 example:
func injectDefaultIfHas(wherePtr any) {
	injector, ok := wherePtr.(InjectDefaulter)
	if ok {
		injector.InjectDefault()
	}
}
