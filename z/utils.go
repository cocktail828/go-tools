package z

import (
	"fmt"
	"strings"

	"github.com/cocktail828/go-tools/z/reflectx"
)

func Must(err error) {
	if !reflectx.IsNil(err) {
		panic(err)
	}
}

// Error type represents list of errors
type Error []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e Error) Error() string {
	if len(e) == 0 {
		return ""
	}
	errstrs := make([]string, len(e))
	for i, l := range e {
		if l != nil {
			errstrs[i] = fmt.Sprintf("#%d: %s", i+1, l.Error())
		}
	}
	return fmt.Sprintf("Errors (num=%v):\n%s", len(e), strings.Join(errstrs, "\n"))
}

func (e Error) Last() error {
	if len(e) == 0 {
		return nil
	}
	return e[len(e)-1]
}
