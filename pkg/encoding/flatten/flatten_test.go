package flatten

import (
	"encoding/json"
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

var (
	jsonData = `{
	"common": {
		"app_id": "xxx",
		"user_id": "yyy",
		"age": 30,
		"height": 175.5,
		"width": 175.5,
		"is_active": true
	},
	"business": {
		"res_ids": ["a", "b"],
		"scores": [90, 85],
		"flags": [true, false],
		"bypass0": {
			"a0": "a",
			"a1": "a"
		},
		"bypass1": {
			"a0": 1,
			"a1": 1
		}
	}
}`
)

type MyStruct struct {
	AppID    *string           `flatten:"common.app_id" json:"common.app_id"`
	UserID   *string           `flatten:"common.user_id" json:"common.user_id"`
	Age      *int              `flatten:"common.age" json:"common.age"`
	Height   *float64          `flatten:"common.height" json:"common.height"`
	Width    *float32          `flatten:"common.width" json:"common.width"`
	IsActive *bool             `flatten:"common.is_active" json:"common.is_active"`
	ResIDs   []string          `flatten:"business.res_ids" json:"business.res_ids"`
	Scores   []int             `flatten:"business.scores" json:"business.scores"`
	Flags    []*bool           `flatten:"business.flags" json:"business.flags"`
	ByPass0  map[string]string `flatten:"business.bypass0" json:"business.bypass0"`
	ByPass1  map[string]int    `flatten:"business.bypass1" json:"business.bypass1"`
}

func TestUnmarshal(t *testing.T) {
	var result MyStruct
	z.Must(Unmarshal([]byte(jsonData), &result))
	assert.Equal(t, "xxx", *result.AppID)
	assert.Equal(t, "yyy", *result.UserID)
	assert.Equal(t, 30, *result.Age)
	assert.Equal(t, float64(175.5), *result.Height)
	assert.Equal(t, float32(175.5), *result.Width)
	assert.Equal(t, true, *result.IsActive)
	assert.Equal(t, []string{"a", "b"}, result.ResIDs)
	assert.Equal(t, []int{90, 85}, result.Scores)
	assert.Equal(t, true, *result.Flags[0])
	assert.Equal(t, false, *result.Flags[1])
	assert.EqualValues(t, map[string]string{"a0": "a", "a1": "a"}, result.ByPass0)
	assert.EqualValues(t, map[string]int{"a0": 1, "a1": 1}, result.ByPass1)
}

func BenchmarkFlatten(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var result MyStruct
			z.Must(Unmarshal([]byte(jsonData), &result))
		}
	})
	b.ReportAllocs()
}

func BenchmarkJson(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			mmp := map[string]any{}
			z.Must(json.Unmarshal([]byte(jsonData), &mmp))
		}
	})
	b.ReportAllocs()
}
