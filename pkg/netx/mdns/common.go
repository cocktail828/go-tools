package mdns

import (
	"io"
	"log"
)

var (
	mdnsLogger = log.New(io.Discard, "mdns: ", log.LstdFlags)
)

func SetLogger(l *log.Logger) {
	if l != nil {
		mdnsLogger = l
	}
}
