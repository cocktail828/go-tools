package z

import (
	"log"
	"os"
	"strings"
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

func init() {
	switch val := strings.ToLower(os.Getenv("MODE")); val {
	case "develop", "release":
		mode = Mode(val)
	default:
		log.Fatalf("env '%v' should be oneof [develop|release]", val)
	}
}

func DevelopMode() bool { return mode == "debug" }
func ReleaseMode() bool { return mode == "release" }
func SetMode(m Mode)    { mode = m }
