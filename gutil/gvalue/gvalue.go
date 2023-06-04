package gvalue

// Zero returns zero value of type.
//
// The zero value is:
//
//   - 0	for numeric types,
//   - false for the boolean type
//   - "" (the empty string) for strings
//   - nil for reference/pointer type
func Zero[T any]() (v T) {
	return
}
