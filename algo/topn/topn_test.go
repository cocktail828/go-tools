package topn_test

import (
	"testing"

	"github.com/cocktail828/go-tools/algo/topn"
	"github.com/stretchr/testify/assert"
)

type SortBy []int16

func (a SortBy) Len() int           { return len(a) }
func (a SortBy) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortBy) Less(i, j int) bool { return a[i] < a[j] }
func TestTopN(t *testing.T) {
	array := SortBy{1, 3, 7, 2, 4, 0, 9}
	topn.TopN(array, len(array))
	assert.Equal(t, SortBy{0, 1, 2, 3, 4, 7, 9}, array)
}
