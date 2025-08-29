package parser

import (
	"fmt"
	"os"
	"testing"
)

func Test_parser_route(t *testing.T) {
	inputs := []string{
		"post /user/login",
		"post /user/login (Login)",
		"post /user/login return (LoginResp)",
		"post /user/login (Login) return (LoginResp)",
		"post /user/login (Login) return",
	}

	for _, in := range inputs {
		t.Logf("%#v", parserRoute(in))
	}
}

func TestParserGo(t *testing.T) {
	dslContent, _ := os.ReadFile("../tests/demo.dsl")
	structs, err := parseStructs(string(dslContent))
	if err != nil {
		panic(err)
	}

	for _, s := range structs {
		fmt.Printf("struct %s {\n", s.Name)
		for _, f := range s.Fields {
			fmt.Printf("  %s %s\n", f.Name, f.Type)
		}
		fmt.Println("}")
	}
}
