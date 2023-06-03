package gslice

// Insert inserts elements vs before position pos, returns a new allocated slice.
// [Negative index] is supported.
//
//   - Insert(x, 0, ...) inserts at the front of the slice
//   - Insert(x, len(x), ...) is equivalent to append(x, ...)
//   - Insert(x, -1, ...) is equivalent to Insert(x, len(x)-1, ...)
//
// 🚀 EXAMPLE:
//
//	s := []int{0, 1, 2, 3}
//	Insert(s, 0, 99)      ⏩ []int{99, 0, 1, 2, 3}
//	Insert(s, 0, 98, 99)  ⏩ []int{98, 99, 0, 1, 2, 3}
//	Insert(s, 4, 99)      ⏩ []int{0, 1, 2, 3, 99}
//	Insert(s, 1, 99)      ⏩ []int{0, 99, 1, 2, 3}
//	Insert(s, -1, 99)     ⏩ []int{0, 1, 2, 99, 3}
func Insert[T any](s []T, pos int, vs ...T) []T {
	if len(vs) == 0 {
		return Clone(s)
	}
	pos, _ = normalizeIndex(s, pos)
	if pos >= len(s) {
		pos = len(s)
	} else if pos < 0 {
		pos = 0
	}

	dst := make([]T, len(s)+len(vs))
	copy(dst, s[:pos])
	copy(dst[pos:], vs)
	copy(dst[pos+len(vs):], s[pos:])
	return dst
}

// Clone returns a shallow copy of the slice.
// If the given slice is nil, nil is returned.
//
// 🚀 EXAMPLE:
//
//	Clone([]int{1, 2, 3}) ⏩ []int{1, 2, 3}
//	Clone([]int{})        ⏩ []int{}
//	Clone[int](nil)       ⏩ nil
//
// 💡 HINT: The elements are copied using assignment (=), so this is a shallow clone.
// If you want to do a deep clone, use [CloneBy] with an appropriate element
// clone function.
//
// 💡 AKA: Copy
func Clone[T any, S ~[]T](s S) S {
	if s == nil {
		return nil
	}
	cloned := make(S, len(s))
	for i := range s {
		cloned[i] = s[i]
	}
	return cloned
}

// ToMap collects elements of slice to map, both map keys and values are produced
// by mapping function f.
//
// 🚀 EXAMPLE:
//
//	type Foo struct {
//		ID   int
//		Name string
//	}
//	mapper := func(f Foo) (int, string) { return f.ID, f.Name }
//	ToMap([]Foo{}, mapper) ⏩ map[int]string{}
//	s := []Foo{{1, "one"}, {2, "two"}, {3, "three"}}
//	ToMap(s, mapper)       ⏩ map[int]string{1: "one", 2: "two", 3: "three"}
func ToMap[T, V any, K comparable](s []T, f func(T) (K, V)) map[K]V {
	m := make(map[K]V, len(s))
	for _, e := range s {
		k, v := f(e)
		m[k] = v
	}
	return m
}

// normalizeIndex normalizes possible [Negative index] to positive index.
// the returned bool indicate whether the normalized index is in range [0, len(s)).
func normalizeIndex[T any](s []T, i int) (int, bool) {
	if i < 0 {
		i += len(s)
	}
	return i, i >= 0 && i < len(s)
}
