package gen

type StringHandler func(string) string

func titleSlice(f StringHandler, ss ...string) []string {
	rs := make([]string, len(ss))
	for i, s := range ss {
		rs[i] = f(s)
	}

	return rs
}
