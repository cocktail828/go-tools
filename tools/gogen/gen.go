package main

import (
	"log"
	"os"

	gogen "github.com/cocktail828/go-tools/tools/gogen/gen"
	"github.com/cocktail828/go-tools/tools/gogen/parser"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: generator <dsl file> <target dir>")
	}

	dslFile := os.Args[1]
	dsl, err := parser.ParseDSL(dslFile)
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}

	rootdir := os.Args[2]
	os.MkdirAll(rootdir, 0755)
	if err := gogen.Generate(rootdir, dsl); err != nil {
		log.Fatalf("generate error: %v", err)
	}

	log.Println("Code generated successfully.")
}
