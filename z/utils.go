package z

import (
	"fmt"

	"github.com/cocktail828/go-tools/z/reflectx"
)

type Mode string

const (
	Development = Mode("develop")
	Release     = Mode("release")
)

var (
	// indicates environment name for work mode
	mode = Development
)

func DevelopMode() bool { return mode == "debug" }
func ReleaseMode() bool { return mode == "release" }
func SetMode(m Mode)    { mode = m }

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
