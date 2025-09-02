package kvstore

import (
	"testing"

	"github.com/cocktail828/go-tools/z/variadic"
)

func TestOption(t *testing.T) {
	c := variadic.Compose(TTL(500), MatchPrefix(), IgnoreLease(), Limit(100), Count())

	t.Log("TTL:", GetTTL(c))                 // 输出: TTL: 500
	t.Log("MatchPrefix:", GetMatchPrefix(c)) // 输出: MatchPrefix: true
	t.Log("IgnoreLease:", GetIgnoreLease(c)) // 输出: IgnoreLease: true
	t.Log("Limit:", GetLimit(c))             // 输出: Limit: 100
	t.Log("Count:", GetCount(c))             // 输出: Count: true
}
