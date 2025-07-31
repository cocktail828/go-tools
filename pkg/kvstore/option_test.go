package kvstore_test

import (
	"testing"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	"github.com/cocktail828/go-tools/z/variadic"
)

func TestOption(t *testing.T) {
	v := kvstore.Variadic([]variadic.Option{
		kvstore.TTL(500),
		kvstore.MatchPrefix(),
		kvstore.IgnoreLease(),
		kvstore.Limit(100),
		kvstore.Count(),
	}...)

	t.Log("TTL:", v.TTL())                 // 输出: TTL: 500
	t.Log("MatchPrefix:", v.MatchPrefix()) // 输出: MatchPrefix: true
	t.Log("IgnoreLease:", v.IgnoreLease()) // 输出: IgnoreLease: true
	t.Log("Limit:", v.Limit())             // 输出: Limit: 100
	t.Log("Count:", v.Count())             // 输出: Count: true
}
