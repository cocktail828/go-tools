package etcdkv

type CountResult struct{ Num int }

func (r CountResult) Len() int           { return r.Num }
func (r CountResult) Key(i int) string   { return "" }
func (r CountResult) Value(i int) []byte { return nil }
