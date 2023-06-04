package gslice

// Of creates a slice from variadic arguments.
// If no argument given, an empty (non-nil) slice []T{} is returned.
//
// 💡 HINT: This function is used to omit verbose types like "[]LooooongTypeName{}"
// when constructing slices.
//
// 🚀 EXAMPLE:
//
//	Of(1, 2, 3) ⏩ []int{1, 2, 3}
//	Of(1)       ⏩ []int{1}
//	Of[int]()   ⏩ []int{}
func Of[T any](v ...T) []T {
	if len(v) == 0 {
		return []T{} // never return nil
	}
	return v
}
