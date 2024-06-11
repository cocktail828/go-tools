package cmd

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	DefaultSignals = []os.Signal{syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT}
)

/*
watchdog 用来作为服务的启动进程存在, 他会再延迟时间后注册服务, 或者提前结束注册
并在感知到服务退出时立刻取消注册, 或者收到信号提前取消注册
如果收到信号退出, 则需要等待指定时间(非SIGCHLD)
*/
type Watchdog struct {
	OnEvent    func(sig os.Signal) // 事件回调
	Register   func()
	DeRegister func()
	inited     atomic.Bool
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
}

func (w *Watchdog) Spawn(name string, args ...string) error {
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.cmd = exec.CommandContext(w.ctx, name, args...)
	w.cmd.Stderr = os.Stderr
	w.cmd.Stdout = os.Stdout
	defer w.cancel()
	w.inited.Store(true)
	if err := w.cmd.Start(); err != nil {
		return err
	}
	return w.cmd.Wait()
}

func (w *Watchdog) WaitForSignal(tmo time.Duration, signals ...os.Signal) {
	if len(signals) == 0 {
		signals = DefaultSignals
	}

	for !w.inited.Load() {
		time.Sleep(time.Millisecond * 10)
	}

	select {
	case <-w.ctx.Done(): // already quit
		return
	default:
	}

	// register service
	if w.Register != nil {
		w.Register()
	}

	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, signals...)
	defer signal.Stop(signalChan)

	once := sync.Once{}
	f := func() {
		once.Do(func() {
			if w.DeRegister != nil {
				w.DeRegister()
			}
		})
	}
	defer f()

	for {
		select {
		case <-w.ctx.Done(): // child exit?
			return

		case sig := <-signalChan: // wait for signals
			w.OnEvent(sig)
			// SIGCHLD does not mean child is exited
			// however, w.ctx.Done() will return if child exit
			if sig != syscall.SIGCHLD {
				f()
				if w.cmd.Process == nil {
					return
				}
				w.cmd.Process.Signal(sig)
				time.AfterFunc(tmo, func() { w.cmd.Process.Kill() })
			}
		}
	}
}
