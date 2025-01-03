package z

func Contains[S ~[]E, E comparable](s S, e E) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
