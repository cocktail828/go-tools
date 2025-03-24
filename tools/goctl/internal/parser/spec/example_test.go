package spec_test

import (
	"fmt"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
)

func ExampleMember_GetEnumOptions() {
	member := spec.Member{
		Tag: `json:"foo,options=foo|bar|options|123"`,
	}
	fmt.Println(member.GetEnumOptions())
	// Output:
	// [foo bar options 123]
}
