package gdal

// InjectDefaulter
//
// @Description: *Where 默认值注入
//
// 查询条件对象 *Where 如果实现了此接口，在查询/更新前会自动调用 DefaultInject()
//
// ❗ 请为 *Where 而不是 Where 实现此接口，因为 Where 的方法使用的是它的拷贝
type InjectDefaulter interface {
	InjectDefault()
}

// injectDefaultIfHas
// @Description: 如果 *Where 指定了默认值，则将其注入进去
// @param wherePtr: 查询条件指针
func injectDefaultIfHas(wherePtr any) {
	injector, ok := wherePtr.(InjectDefaulter)
	if ok {
		injector.InjectDefault()
	}
}
