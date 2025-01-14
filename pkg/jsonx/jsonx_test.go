package jsonx_test

import (
	"encoding/json"
	"testing"

	"github.com/cocktail828/go-tools/pkg/jsonx"
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
	AppID    *string           `jsonx:"common.app_id"`
	UserID   *string           `jsonx:"common.user_id"`
	Age      *int              `jsonx:"common.age"`
	Height   *float64          `jsonx:"common.height"`
	Width    *float32          `jsonx:"common.width"`
	IsActive *bool             `jsonx:"common.is_active"`
	ResIDs   []string          `jsonx:"business.res_ids"`
	Scores   []int             `jsonx:"business.scores"`
	Flags    []*bool           `jsonx:"business.flags"`
	ByPass0  map[string]string `jsonx:"business.bypass0"`
	ByPass1  map[string]int    `jsonx:"business.bypass1"`
}

func TestUnmarshal(t *testing.T) {
	var result MyStruct
	z.Must(jsonx.Unmarshal([]byte(jsonData), &result))
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

func BenchmarkJsonx(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var result MyStruct
			z.Must(jsonx.Unmarshal([]byte(jsonData), &result))
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
