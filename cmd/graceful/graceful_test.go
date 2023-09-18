package graceful_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/cmd/graceful"
	"github.com/sirupsen/logrus"
)

func TestGraceful(t *testing.T) {
	g := graceful.New(graceful.Config{
		InitPostPone: time.Second * 3,
		QuitPostPone: time.Second * 3,
		PreStart:     func(ctx context.Context) { logrus.Println("prestart") },
		PostStart:    func(ctx context.Context) { logrus.Println("poststart") },
		PreStop:      func(ctx context.Context) { logrus.Println("prestop") },
		PostStop:     func(ctx context.Context) { logrus.Println("poststop") },
		OnEvent:      func(sig os.Signal) { logrus.Println(sig) },
	})
	logrus.Println(g.Spawn("sleep", "10"))
}
