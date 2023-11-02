package memkey_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/memkey"
)

func TestXXX(t *testing.T) {
	m := memkey.New()
	m.Add("a", func() bool { fmt.Println("a", time.Now()); time.Sleep(time.Millisecond);return true })
	time.Sleep(time.Hour)
}
