package z

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMemory1(t *testing.T) {
	testCases := []string{
		"10B",
		"512KB",
		"100MB",
		"1.5GB",
		"2TB",
		"5PB",

		"10 B",
		"512 KB",
		"100 MB",
		"1.5 GB",
		"2 TB",
		"5 PB",
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

func TestMemory2(t *testing.T) {
	type OBJ struct {
		Memory Memory `json:"memory"`
	}
	num := 10 * TeraByte

	bytes, err := json.Marshal(OBJ{Memory: num})
	if err != nil {
		t.Fatalf("ERROR: %v\n", err)
	} else {
		t.Logf("%s\n", bytes)
	}

	obj := OBJ{}
	if err := json.Unmarshal(bytes, &obj); err != nil {
		t.Fatalf("ERROR: %v\n", err)
	}
	assert.Equal(t, OBJ{Memory: num}, obj)
}
