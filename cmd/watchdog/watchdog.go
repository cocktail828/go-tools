package watchdog

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/cocktail828/go-tools/cmd/call"
)

type Watchdog struct {
	InitPostPone       time.Duration                   // 延迟注册的时间
	QuitPostPone       time.Duration                   // 延迟退出的时间
	OperateTimeout     time.Duration                   // 回调函数超时时间
	InterceptorSignals []os.Signal                     // 优雅启停需要关注的信号, 默认为 syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD
	Register           func(ctx context.Context) error // 注册服务
	DeRegister         func(ctx context.Context) error // 服务去注册
	OnEvent            func(sig os.Signal)             // 事件回调
}

func (g *Watchdog) normalize() {
	if g.InitPostPone == 0 {
		g.InitPostPone = time.Second * 10
	}
	if g.QuitPostPone == 0 {
		g.QuitPostPone = time.Second * 5
	}
	if g.OperateTimeout == 0 {
		g.OperateTimeout = time.Second * 3
	}
	if g.Register == nil {
		g.Register = func(ctx context.Context) error { return nil }
	}
	if g.DeRegister == nil {
		g.DeRegister = func(ctx context.Context) error { return nil }
	}
	if g.OnEvent == nil {
		g.OnEvent = func(sig os.Signal) {}
	}
	if len(g.InterceptorSignals) == 0 {
		g.InterceptorSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD}
	}
}

func (g *Watchdog) Spawn(name string, args ...string) error {
	g.normalize()
	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, g.InterceptorSignals...)
	signal.Ignore(syscall.SIGURG, syscall.SIGWINCH)
	defer signal.Stop(signalChan)

	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	regCancel, regErrChan := call.Delayed(
		g.InitPostPone,
		func(ctx context.Context) error { return call.Timed(g.OperateTimeout, g.Register) },
	)

	for {
		sig := <-signalChan
		regCancel()
		g.OnEvent(sig)
		if sig != syscall.SIGCHLD && cmd != nil && cmd.Process != nil {
			if err := cmd.Process.Signal(sig); err == os.ErrProcessDone {
				break
			}
		}
		if sig == syscall.SIGCHLD {
			break
		}
	}

	if err := <-regErrChan; err == nil {
		_, postStopErrChan := call.Delayed(g.QuitPostPone,
			func(ctx context.Context) error { return call.Timed(g.OperateTimeout, g.DeRegister) },
		)
		<-postStopErrChan
	}
	return cmd.Wait()
}
