package router_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/algo/router"
)

func TestRouter(t *testing.T) {
	r := router.New()
	r.Register("/asd/:id/:a", nil)
	r.Register("/asd/:id", nil)
	fmt.Println(r.Lookup("/asd/1/2"))
	fmt.Println(r.Lookup("/asd/1/"))
	fmt.Println(r.Lookup("/asd1/1"))
}
