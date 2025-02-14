package lumberjack_test

import (
	"os"
	"testing"

	"github.com/cocktail828/go-tools/configor"
	"github.com/cocktail828/go-tools/xlog/lumberjack"
	"github.com/cocktail828/go-tools/z"
)

func BenchmarkLumberjack(b *testing.B) {
	cfg := lumberjack.Config{}
	z.Must(configor.Load(&cfg, []byte(`
level = "INFO"
filename = "/log/server/xxx.log"
async = true
`)))

	os.RemoveAll("/log/server/xxx.log")
	b.Run("no-cache", func(b *testing.B) {
		l := lumberjack.NewWriter(cfg)
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				l.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
			}
		})
	})
}
