package z

import (
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
	a(t)
}

func c(t *testing.T) { t.Log(Stack(5)) }
func b(t *testing.T) { c(t) }
func a(t *testing.T) { b(t) }

func TestSize(t *testing.T) {
	testCases := []string{
		"10B",
		"512KB",
		"1.5GB",
		"2TB",
		"100MB",
	}

	for _, tc := range testCases {
		bytes, err := ParseMemory(tc)
		if err != nil {
			t.Fatalf("%-10s → ERROR: %v\n", tc, err)
		} else {
			t.Logf("%-10s → %v\n", tc, bytes)
		}
	}
}
