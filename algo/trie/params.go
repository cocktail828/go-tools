package trie

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// add appends a new parameter to the slice.
func (ps *Params) add(name, value string) {
	*ps = append(*ps, Param{name, value})
}

// pop removes the last parameter from the slice.
func (ps *Params) pop() {
	if len(*ps) == 0 {
		return
	}
	*ps = (*ps)[:len(*ps)-1]
}
