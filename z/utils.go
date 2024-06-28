package z

import (
	"fmt"
	"log"
	"os"

	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/stretchr/testify/assert"
)

func Must(err error) {
	if !reflectx.IsNil(err) {
		panic(err)
	}
}

func Mustf(err error, format string, args ...any) {
	if !reflectx.IsNil(err) {
		panic(fmt.Sprintf(format, args...))
	}
}

type errorf struct{}

func (e errorf) Errorf(format string, args ...interface{}) { log.Fatalf(format, args...) }

func Assert(cond bool) {
	if ReleaseMode() {
		return
	}
	if !assert.True(errorf{}, cond) {
		os.Exit(1)
	}
}

func Assertf(cond bool, format string, args ...interface{}) {
	if ReleaseMode() {
		return
	}
	if !assert.Truef(errorf{}, cond, format, args...) {
		os.Exit(1)
	}
}
