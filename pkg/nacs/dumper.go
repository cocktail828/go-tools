package nacs

import (
	"os"
	"path/filepath"
)

type dumper struct {
	dir    string
	prefix string
}

func (d dumper) Dump(key string, payload []byte) {
	os.MkdirAll(d.dir, os.ModePerm)
	os.WriteFile(filepath.Join(d.dir, d.prefix+key+".cache"), payload, os.ModePerm)
}

func (d dumper) Remove(key string) {
	os.MkdirAll(d.dir, os.ModePerm)
	os.Remove(filepath.Join(d.dir, d.prefix+key+".cache"))
}
