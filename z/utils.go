package z

import (
	"log"

	"github.com/cocktail828/go-tools/z/reflectx"
)

func Must(err ...error) {
	for _, e := range err {
		if !reflectx.IsNil(e) {
			panic(e)
		}
	}
}

func Assert(cond bool) {
	if ReleaseMode() {
		return
	}
	if !cond {
		log.Fatalf("assert fail")
	}
}

func Assertf(cond bool, format string, args ...any) {
	if ReleaseMode() {
		return
	}
	if !cond {
		log.Fatalf("assert fail, "+format, args...)
	}
}
