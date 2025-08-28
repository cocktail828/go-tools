package parser

import (
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
