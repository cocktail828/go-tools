package graceful

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
	gracefulStopSignals   = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD}
	defaultOperateTimeout = time.Second * 3
	initPostPone          = time.Second * 10
	quitPostPone          = time.Second * 3
)

type Config struct {
	InitPostPone      time.Duration             // 延迟注册的时间
	QuitPostPone      time.Duration             // 延迟退出的时间
	OperateTimeout    time.Duration             // 回调函数超时时间
	InterceporSignals []os.Signal               // 优雅启停需要关注的信号, 默认为 syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD
	PreStart          func(ctx context.Context) // 准备工作
	PostStart         func(ctx context.Context) // 注册服务
	PreStop           func(ctx context.Context) // 服务去注册, 销毁任何 PostStart 创建的资源
	PostStop          func(ctx context.Context) // 销毁环境, 销毁任何 PreStart 创建的资源
	OnEvent           func(sig os.Signal)       // 事件回调
}

func (cfg *Config) normalize() {
	if cfg.InitPostPone == 0 {
		cfg.InitPostPone = initPostPone
	}

	if cfg.QuitPostPone == 0 {
		cfg.QuitPostPone = quitPostPone
	}

	if cfg.OperateTimeout == 0 {
		cfg.OperateTimeout = defaultOperateTimeout
	}

	if cfg.PreStart == nil {
		cfg.PreStart = func(ctx context.Context) {}
	}

	if cfg.PostStart == nil {
		cfg.PostStart = func(ctx context.Context) {}
	}

	if cfg.PreStop == nil {
		cfg.PreStop = func(ctx context.Context) {}
	}

	if cfg.PostStop == nil {
		cfg.PostStop = func(ctx context.Context) {}
	}

	if cfg.OnEvent == nil {
		cfg.OnEvent = func(sig os.Signal) {}
	}

	cfg.InterceporSignals = append(cfg.InterceporSignals, gracefulStopSignals...)
}

type graceful struct {
	cfg        Config
	ctx        context.Context
	cancel     context.CancelFunc
	signalChan chan os.Signal
	cmd        *exec.Cmd
	delayed    sync.Map
}

func WithContext(ctx context.Context, cfg Config) *graceful {
	ctx, cancel := context.WithCancel(ctx)
	g := &graceful{
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		signalChan: make(chan os.Signal),
	}

	g.cfg.normalize()
	signal.Notify(g.signalChan)
	signal.Ignore(syscall.SIGURG, syscall.SIGWINCH)

	return g
}

func (g *graceful) Stop() {
	g.cancel()
}

func (g *graceful) Spawn(name string, args ...string) error {
	g.call(g.cfg.PreStart)
	defer g.call(g.cfg.PostStop)

	g.cmd = exec.CommandContext(g.ctx, name, args...)
	if err := g.cmd.Start(); err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	g.startEventLoop(wg)

	g.delayedCall(g.cfg.InitPostPone, g.cfg.PostStart)
	err := g.cmd.Wait()
	wg.Wait()

	return err
}

func (g *graceful) startEventLoop(wg *sync.WaitGroup) {
	ch := make(chan struct{})
	go func() {
		close(ch)
		defer wg.Done()
		var preStoppCnt atomic.Bool

		for {
			select {
			case sig := <-g.signalChan:
				g.cfg.OnEvent(sig)
				g.cancelDelayed()
				switch sig {
				case syscall.SIGCHLD:
					if preStoppCnt.CompareAndSwap(false, true) {
						g.delayedCall(g.cfg.QuitPostPone, g.cfg.PreStop)
					}
					return

				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					if preStoppCnt.CompareAndSwap(false, true) {
						g.delayedCall(g.cfg.QuitPostPone, g.cfg.PreStop)
					}
				}

				if g.cmd != nil && g.cmd.Process != nil {
					if err := g.cmd.Process.Signal(sig); err == os.ErrProcessDone {
						return
					}
				}

			case <-g.ctx.Done():
				g.cfg.OnEvent(nil)
				g.cancelDelayed()
				if preStoppCnt.CompareAndSwap(false, true) {
					g.delayedCall(g.cfg.QuitPostPone, g.cfg.PreStop)
				}
			}
		}
	}()
	<-ch
}

func (g *graceful) call(f func(context.Context)) {
	ctx, cancel := context.WithTimeout(g.ctx, g.cfg.OperateTimeout)
	defer cancel()
	f(ctx)
}

func (g *graceful) delayedCall(delay time.Duration, f func(context.Context)) {
	cancelCtx, cancelF := context.WithCancel(g.ctx)
	g.delayed.Store(cancelCtx, cancelF)

	select {
	case <-time.After(delay):
		ctx, cancel := context.WithTimeout(cancelCtx, g.cfg.OperateTimeout)
		defer cancel()
		f(ctx)
	case <-cancelCtx.Done():
	}
}

func (g *graceful) cancelDelayed() {
	g.delayed.Range(func(key, value any) bool {
		value.(context.CancelFunc)()
		return true
	})
}
