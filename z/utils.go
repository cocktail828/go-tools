package z

import (
	"github.com/cocktail828/go-tools/z/reflectx"
)

func Must(err error) {
	if !reflectx.IsNil(err) {
		panic(err)
	}
}
