package gptr

// Of returns a pointer that points to equivalent value of value v.
// (T → *T).
// It is useful when you want to "convert" a unaddressable value to pointer.
//
// If you need to assign the address of a literal to a pointer:
//
//	 payload := struct {
//		    Name *string
//	 }
//
// The practice without generic:
//
//	x := "name"
//	payload.Name = &x
//
// Use generic:
//
//	payload.Name = Of("name")
//
// 💡 HINT: use [Indirect] to dereference pointer (*T → T).
//
// ⚠️  WARNING: The returned pointer does not point to the original value because
// Go is always pass by value, user CAN NOT modify the value by modifying the pointer.
func Of[T any](v T) *T {
	return &v
}
