package graceful_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/graceful"
)

func TestWatchdog(t *testing.T) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	w := graceful.Watchdog{
		Postpone:   time.Second*3,
		Register:   func() { fmt.Println(time.Now(), "reg") },
		DeRegister: func() { fmt.Println(time.Now(), "dreg") },
	}
	log.Println(time.Now())
	log.Println(w.Respawn(c, "sleep", "1"))
	log.Println(time.Now())
}
