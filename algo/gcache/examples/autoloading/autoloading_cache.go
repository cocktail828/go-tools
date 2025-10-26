package main

import (
	"fmt"

	"github.com/cocktail828/go-tools/algo/gcache"
)

func main() {
	gc := gcache.New(10).LFU().
		LoaderFunc(func(key any) (any, error) {
			return fmt.Sprintf("%v-value", key), nil
		}).Build()

	v, err := gc.Get("key")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
