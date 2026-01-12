package snowflake

import (
	"testing"

	"github.com/cocktail828/go-tools/pkg/setx"
	"github.com/cocktail828/go-tools/z"
)

func TestGenerateDuplicateID(t *testing.T) {
	node, err := NewNode(1)
	z.Must(err)

	set := setx.NewSet[int64]()
	for i := 0; i < 1000000; i++ {
		y := node.Generate()
		if set.Contains(y) {
			t.Errorf("y(%d) is duplicated", y)
		}
		set.Add(y)
	}
}
