package inet

import (
	"log"
	"testing"
)

func TestInet(t *testing.T) {
	log.Println(Inet4())
	log.Println(Inet6())
}
