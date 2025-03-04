package main

import (
	"github.com/cocktail828/go-tools/xlog/assert"
)

func main() {
	assert.Println("this line wont shown")

	assert.SetDebugMode()
	assert.Println("this line will show")

	assert.SetReleaseMode()
	assert.Println("this line wont shown")
}
