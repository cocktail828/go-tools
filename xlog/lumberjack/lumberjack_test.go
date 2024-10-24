package lumberjack_test

import (
	"os"
	"testing"

	"github.com/cocktail828/go-tools/configor"
	"github.com/cocktail828/go-tools/xlog/lumberjack"
	"github.com/cocktail828/go-tools/z"
)

func BenchmarkLog(b *testing.B) {
	cfg := lumberjack.Config{}
	z.Must(configor.Load(&cfg, []byte(`
level = "INFO"
filename = "/log/server/xxx.log"
async = true
`)))

	os.RemoveAll("/log/server/*")
	b.Run("no-cache", func(b *testing.B) {
		cfg.Async = false
		l := lumberjack.NewLumberjack(cfg)
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				l.Errorln("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
			}
		})
	})

	os.RemoveAll("/log/server/*")
	b.Run("cache", func(b *testing.B) {
		cfg.Async = true
		l := lumberjack.NewLumberjack(cfg)
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				l.Errorln("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
			}
		})
	})
}
