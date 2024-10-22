package z

import (
	"fmt"

	"github.com/cocktail828/go-tools/z/reflectx"
)

type Mode string

const (
	Debug   = Mode("debug")
	Test    = Mode("test")
	Release = Mode("release")
)

var (
	// indicates environment name for work mode
	mode = Debug
)

func DebugMode() bool   { return mode == Debug }
func TestMode() bool    { return mode == Test }
func ReleaseMode() bool { return mode == Release }
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
