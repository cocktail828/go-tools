package snowflake

import (
	"testing"
)

func TestGenerateDuplicateID(t *testing.T) {
	var x, y int64
	node, _ := NewNode(1)
	for i := 0; i < 1000000; i++ {
		y = node.Generate()
		if x == y {
			t.Errorf("x(%d) & y(%d) are the same", x, y)
		}
		x = y
	}
}
