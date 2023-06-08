package graceful

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
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
	InitPostPone      time.Duration // 延迟注册的时间
	QuitPostPone      time.Duration // 延迟退出的时间
	OperateTimeout    time.Duration // 回调函数超时时间
	InterceporSignals []os.Signal   // 优雅启停需要关注的信号, 默认为 syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD
	PreStart          func(ctx context.Context) error
	PostStart         func(ctx context.Context) error
	PreStop           func(ctx context.Context) error
	PostStop          func(ctx context.Context) error
}

type graceful struct {
	initPostPone      time.Duration
	quitPostPone      time.Duration
	operateTimeout    time.Duration
	interceporSignals []os.Signal
	preStart          func() error
	postStart         func() error
	preStop           func() error
	postStop          func() error
	cmdctx            context.Context
	cancel            context.CancelFunc
	signalChan        chan os.Signal
	cmd               *exec.Cmd
	wg                sync.WaitGroup
	cancelFuncs       sync.Map
	once              sync.Once
	onceErr           error
}

func WithContext(ctx context.Context, cfg Config) *graceful {
	ctx, cancel := context.WithCancel(ctx)
	g := &graceful{
		initPostPone:      cfg.InitPostPone,
		quitPostPone:      cfg.QuitPostPone,
		operateTimeout:    cfg.OperateTimeout,
		interceporSignals: cfg.InterceporSignals,
		cmdctx:            ctx,
		cancel:            cancel,
		signalChan:        make(chan os.Signal),
	}

	if g.operateTimeout == 0 {
		g.operateTimeout = defaultOperateTimeout
	}

	if g.initPostPone == 0 {
		g.initPostPone = initPostPone
	}

	if g.quitPostPone == 0 {
		g.quitPostPone = quitPostPone
	}

	g.interceporSignals = append(g.interceporSignals, gracefulStopSignals...)

	g.preStart = func() error {
		if f := cfg.PreStart; f != nil {
			ctx, cancel := context.WithTimeout(g.cmdctx, g.operateTimeout)
			defer cancel()
			return f(ctx)
		}
		return nil
	}

	g.postStart = func() error {
		if f := cfg.PostStart; f != nil {
			ctx, cancel := context.WithTimeout(g.cmdctx, g.initPostPone)
			g.cancelFuncs.Store(ctx, cancel)

			select {
			case <-time.After(g.initPostPone):
				ctx, cancel := context.WithTimeout(g.cmdctx, g.operateTimeout)
				defer cancel()
				return f(ctx)

			case <-ctx.Done():
				return nil
			}
		}
		return nil
	}

	g.preStop = func() error {
		if f := cfg.PreStop; f != nil {
			ctx, cancel := context.WithTimeout(g.cmdctx, g.operateTimeout)
			g.cancelFuncs.Store(ctx, cancel)
			g.once.Do(func() {
				g.onceErr = f(ctx)
				<-time.After(g.quitPostPone)
			})
			return g.onceErr
		}
		return nil
	}

	g.postStop = func() error {
		if f := cfg.PostStop; f != nil {
			ctx, cancel := context.WithTimeout(g.cmdctx, g.operateTimeout)
			defer cancel()
			return f(ctx)
		}
		return nil
	}

	signal.Notify(g.signalChan)
	signal.Ignore(syscall.SIGURG, syscall.SIGWINCH)

	return g
}

func (g *graceful) Start(name string, args ...string) error {
	if err := g.preStart(); err != nil {
		return err
	}

	g.cmd = exec.CommandContext(g.cmdctx, name, args...)
	if err := g.cmd.Start(); err != nil {
		return err
	}

	g.wg.Add(1)
	g.startEventLoop()
	return g.postStart()
}

func (g *graceful) Wait() error {
	var err error
	if g.cmd != nil {
		err = g.cmd.Wait()
	}
	g.wg.Wait()

	return err
}

func (g *graceful) Stop() {
	g.cancel()
}

func (g *graceful) Signal(sig os.Signal) error {
	if g.cmd != nil && g.cmd.Process != nil {
		return g.cmd.Process.Signal(sig)
	}
	return nil
}

func (g *graceful) startEventLoop() {
	ch := make(chan struct{}, 1)
	go func() {
		defer g.wg.Done()
		defer g.postStop()
		ch <- struct{}{}

		for {
			select {
			case sig := <-g.signalChan:
				g.cancelHooks()
				switch sig {
				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD:
					g.preStop()
					if sig == syscall.SIGCHLD {
						return
					}
				}

				if g.cmd != nil && g.cmd.Process != nil {
					if err := g.cmd.Process.Signal(sig); err == os.ErrProcessDone {
						return
					}
				}

			case <-g.cmdctx.Done():
				g.preStop()
				return
			}
		}
	}()
	<-ch
}

func (g *graceful) cancelHooks() {
	g.cancelFuncs.Range(func(key, value any) bool {
		value.(context.CancelFunc)()
		return true
	})
}
