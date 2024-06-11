package cmd_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/cmd"
)

func TestWatchdog(t *testing.T) {
	w := cmd.Watchdog{
		OnEvent:    func(sig os.Signal) { log.Println("signal", sig) },
		Register:   func() { log.Println("reg") },
		DeRegister: func() { log.Println("dereg") },
	}
	go func() {
		log.Println("cmd=>", w.Spawn("sleep", "3"))
	}()
	log.Println(time.Now())
	w.WaitForSignal(time.Second*3, cmd.DefaultSignals...)
}
