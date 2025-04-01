package stringx

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	for i := 0; i < 3; i++ {
		fmt.Println(RandomName(WithCase(), WithWidth(20)))
	}
}
