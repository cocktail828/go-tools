package z

import (
	"github.com/cocktail828/go-tools/z/environ"
)

type Mode string

const (
	Debug   Mode = "debug"
	Test    Mode = "test"
	Release Mode = "release"
)

var (
	// indicates environment name for work mode
	mode = Mode(environ.String("MODE", environ.WithString(string(Release))))
)

func SetMode(m Mode) { mode = m }
func GetMode() Mode  { return mode }
