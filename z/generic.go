package z

func ToMap[T comparable](collection []T) map[T]struct{} {
	tmp := make(map[T]struct{}, len(collection))
	for _, e := range collection {
		tmp[e] = struct{}{}
	}
	return tmp
}

func ToSlice[T comparable](collection map[T]struct{}) []T {
	tmp := make([]T, 0, len(collection))
	for e := range collection {
		tmp = append(tmp, e)
	}
	return tmp
}

func Contains[T comparable](collection []T, eles ...T) bool {
	tmp := ToMap[T](collection)
	for _, e := range eles {
		if _, ok := tmp[e]; !ok {
			return false
		}
	}
	return true
}

func Unique[T comparable](collection []T) []T {
	tmp := map[T]struct{}{}
	r := make([]T, 0, len(collection))
	for _, e := range collection {
		if _, has := tmp[e]; !has {
			tmp[e] = struct{}{}
			r = append(r, e)
		}
	}
	return r
}

func Equal[T comparable](A, B []T) bool {
	if len(A) != len(B) {
		return false
	}
	return Contains[T](A, B...)
}
