package z

import (
	"fmt"
	"slices"
	"testing"
)

func TestZ(t *testing.T) {
	i := 0
	for {
		v := []int{1, 2, 3, 4, 5, 6}
		t.Log(slices.Delete(v, i, i+1))
		i++
		if i >= len(v) {
			break
		}
	}
	a()
}

func c() { fmt.Println(Stack(5)) }
func b() { c() }
func a() { b() }

func TestVersion(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"0.1", "v0.1"},
		{"v0.1", "v0.1"},
		{"v0.1-rc0", "v0.1"},

		{"0.0.1", "v0.0.1"},
		{"v0.0.1", "v0.0.1"},
		{"v0.0.1-rc0", "v0.0.1"},

		{"0.0.1.1", "v0.0.1"},
		{"v0.0.1.1", "v0.0.1"},
		{"v0.0.1.1-rc0", "v0.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := Version(tt.in); got != tt.out {
				t.Errorf("Version() = %v, want %v", got, tt.out)
			}
		})
	}
}
