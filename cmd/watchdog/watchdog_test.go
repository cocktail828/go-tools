package watchdog_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/cmd/watchdog"
)

func TestGraceful(t *testing.T) {
	w := watchdog.Watchdog{
		InitPostPone: time.Second * 5,
		PreStart:     func(ctx context.Context) error { log.Println("prestart"); return nil },
		PostStart:    func(ctx context.Context) error { log.Println("poststart"); return nil },
		PreStop:      func(ctx context.Context) error { log.Println("prestop"); return nil },
		PostStop:     func(ctx context.Context) error { log.Println("poststop"); return nil },
		OnEvent:      func(sig os.Signal) { log.Println("signal", sig) },
	}
	log.Println("cmd", w.Spawn("sleep", "10"))
}
