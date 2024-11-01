package router_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/router"
)

func TestRouter(t *testing.T) {
	r := router.New()
	r.Register("/asd/:id/:a", nil)
	r.Register("/asd/:id", nil)
	fmt.Println(r.Lookup("/asd/1/2"))
	fmt.Println(r.Lookup("/asd/1/"))
	fmt.Println(r.Lookup("/asd/1"))
	fmt.Println(r.Lookup("/asd1/1"))
}

func BenchmarkLookup(b *testing.B) {
	r := router.New()
	r.Register("/asd/:id/:a", nil)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_, ps := r.Lookup("/asd/1/2")
			if len(ps) != 2 || ps.ByName("id") != "1" || ps.ByName("a") != "2" {
				panic(ps)
			}
			_, ps = r.Lookup("/asd/1/")
			if len(ps) != 1 || ps.ByName("id") != "1" {
				panic(ps)
			}
		}
	})
}
