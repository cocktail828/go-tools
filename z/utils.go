package z

import (
	"github.com/cocktail828/go-tools/z/reflectx"
)

func Must(err ...error) {
	for _, e := range err {
		if !reflectx.IsNil(e) {
			panic(e)
		}
	}
}
