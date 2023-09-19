package graceful

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cocktail828/go-tools/cmd"
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
		cfg.InitPostPone = time.Second * 10
	}

	if cfg.QuitPostPone == 0 {
		cfg.QuitPostPone = time.Second * 3
	}

	if cfg.OperateTimeout == 0 {
		cfg.OperateTimeout = time.Second * 3
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

	cfg.InterceporSignals = append(cfg.InterceporSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD)
}

type graceful struct {
	cfg        Config
	ctx        context.Context
	cancel     context.CancelFunc
	signalChan chan os.Signal
	cmd        *exec.Cmd
	once       sync.Once
}

func New(cfg Config) *graceful {
	ctx, cancel := context.WithCancel(context.Background())
	g := &graceful{
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		signalChan: make(chan os.Signal, 3),
	}

	g.cfg.normalize()
	signal.Notify(g.signalChan)
	signal.Ignore(syscall.SIGURG, syscall.SIGWINCH)

	return g
}

func (g *graceful) Stop() {
	g.cancel()
}

// this should only be call once
func (g *graceful) Spawn(name string, args ...string) error {
	cmd.TimedContext(g.ctx, g.cfg.OperateTimeout, g.cfg.PreStart)
	defer func() {
		cancelCtx, _ := cmd.DelayedContext(g.ctx, g.cfg.QuitPostPone, func(ctx context.Context) {
			cmd.TimedContext(g.ctx, g.cfg.OperateTimeout, g.cfg.PostStop)
		})
		<-cancelCtx.Done()
	}()

	postStartCalled := false
	defer func() {
		if !postStartCalled {
			g.Stop()
		}
	}()

	g.cmd = exec.CommandContext(g.ctx, name, args...)
	if err := g.cmd.Start(); err != nil {
		return err
	}

	_, delayCancel := cmd.DelayedContext(g.ctx, g.cfg.InitPostPone, func(ctx context.Context) {
		postStartCalled = true
		cmd.TimedContext(g.ctx, g.cfg.OperateTimeout, g.cfg.PostStart)
	})
	cmd.Async(func() {
		defer delayCancel()
		for {
			select {
			case sig := <-g.signalChan:
				delayCancel()
				g.cfg.OnEvent(sig)
				g.once.Do(func() { cmd.TimedContext(g.ctx, g.cfg.OperateTimeout, g.cfg.PreStop) })
				if g.cmd != nil && g.cmd.Process != nil {
					if err := g.cmd.Process.Signal(sig); err == os.ErrProcessDone {
						return
					}
				}

			case <-g.ctx.Done():
				delayCancel()
				g.cfg.OnEvent(nil)
				g.once.Do(func() { cmd.TimedContext(g.ctx, g.cfg.OperateTimeout, g.cfg.PreStop) })
			}
		}
	})
	return g.cmd.Wait()
}
