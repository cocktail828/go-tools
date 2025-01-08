package kvstore_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/pkg/kvstore"
)

func TestOption(t *testing.T) {
	v := kvstore.Variadic([]kvstore.Option{
		kvstore.TTL(500),
		kvstore.MatchPrefix(),
		kvstore.IgnoreLease(),
		kvstore.Limit(100),
		kvstore.Count(),
	}...)

	fmt.Println("TTL:", v.TTL())                 // 输出: TTL: 500
	fmt.Println("MatchPrefix:", v.MatchPrefix()) // 输出: MatchPrefix: true
	fmt.Println("IgnoreLease:", v.IgnoreLease()) // 输出: IgnoreLease: true
	fmt.Println("Limit:", v.Limit())             // 输出: Limit: 100
	fmt.Println("Count:", v.Count())             // 输出: Count: true
}
