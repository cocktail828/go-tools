package z

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/cocktail828/go-tools/z/stringx"
)

var (
	testRegexp = regexp.MustCompile(`_test|(\\.test$)`)
	// indicates environment name for work mode
	mode = "debug"
)

func init() {
	if testRegexp.MatchString(os.Args[0]) {
		mode = "test"
	} else {
		name := "MODE"
		if val := strings.ToLower(os.Getenv(name)); val != "" {
			if !stringx.Oneof(val, "debug", "release", "test") {
				log.Fatalf("env '%v' should be oneof [debug|release|test]", name)
			}
			mode = val
		}
	}
}

func DebugMode() bool   { return mode == "debug" }
func ReleaseMode() bool { return mode == "release" }
func TestMode() bool    { return mode == "test" }
