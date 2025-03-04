package assert

import (
	"strings"

	"github.com/cocktail828/go-tools/z/environ"
)

const (
	_Debug   = "debug"
	_Test    = "test"
	_Release = "release"
)

var (
	// indicates environment name for work mode
	mode = strings.ToLower(environ.String("MODE", environ.WithString(_Release)))
)

func IsDebugMode() bool   { return mode == _Debug }
func IsTestMode() bool    { return mode == _Test }
func IsReleaseMode() bool { return mode != _Debug && mode != _Test }

func SetDebugMode()   { mode = _Debug }
func SetTestMode()    { mode = _Test }
func SetReleaseMode() { mode = _Release }
