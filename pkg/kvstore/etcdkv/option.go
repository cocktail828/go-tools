package etcdkv

type CountResult struct{ Num int }

func (r CountResult) Len() int         { return r.Num }
func (r CountResult) Key(int) string   { return "" }
func (r CountResult) Value(int) []byte { return nil }
