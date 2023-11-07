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
		QuitPostPone: time.Second * 3,
		Register:     func(ctx context.Context) error { log.Println("Register"); return nil },
		DeRegister:   func(ctx context.Context) error { log.Println("DeRegister"); return nil },
		OnEvent:      func(sig os.Signal) { log.Println("signal", sig) },
	}
	log.Println("cmd", w.Spawn("sleep", "10"))
}
