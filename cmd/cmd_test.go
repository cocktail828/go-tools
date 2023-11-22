package cmd_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/cmd"
)

func TestWatchdog(t *testing.T) {
	w := cmd.Watchdog{
		InitPostPone: time.Second * 5,
		QuitPostPone: time.Second * 3,
		Register:     func(ctx context.Context) { log.Println("Register") },
		DeRegister:   func(ctx context.Context) { log.Println("DeRegister") },
		OnEvent:      func(sig os.Signal) { log.Println("signal", sig) },
	}
	fmt.Println(time.Now())
	log.Println("cmd", w.Spawn("sleep", "10"))
	fmt.Println(time.Now())
}
