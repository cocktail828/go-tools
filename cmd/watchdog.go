package cmd

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

/*
watchdog 用来作为服务的启动进程存在, 他会再延迟时间后注册服务, 或者提前结束注册
并在感知到服务退出时立刻取消注册, 或者收到信号提前取消注册
如果收到信号退出, 则需要等待指定时间(非SIGCHLD)
*/
type Watchdog struct {
	InitPostPone       time.Duration             // 延迟注册的时间
	QuitPostPone       time.Duration             // 延迟退出的时间
	OperateTimeout     time.Duration             // 回调函数超时时间
	InterceptorSignals []os.Signal               // 优雅启停需要关注的信号, 默认为 syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD
	Register           func(ctx context.Context) // 注册服务
	DeRegister         func(ctx context.Context) // 服务去注册
	OnEvent            func(sig os.Signal)       // 事件回调
}

func (w *Watchdog) normalize() {
	if w.InitPostPone == 0 {
		w.InitPostPone = time.Second * 10
	}
	if w.QuitPostPone == 0 {
		w.QuitPostPone = time.Second * 5
	}
	if w.OperateTimeout == 0 {
		w.OperateTimeout = time.Second * 3
	}
	if w.OnEvent == nil {
		w.OnEvent = func(sig os.Signal) {}
	}
	if len(w.InterceptorSignals) == 0 {
		w.InterceptorSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD}
	}
}

func (w *Watchdog) asyncReg() context.CancelFunc {
	regCtx, regCancel := context.WithTimeout(context.Background(), w.OperateTimeout+w.InitPostPone)
	go func() {
		defer regCancel()
		select {
		case <-time.After(w.InitPostPone):
			w.Register(regCtx)
		case <-regCtx.Done():
		}
	}()
	return regCancel
}

func (w *Watchdog) syncDeReg() {
	deRegCtx, deRegCancel := context.WithTimeout(context.Background(), w.OperateTimeout)
	defer deRegCancel()
	if w.DeRegister != nil {
		w.DeRegister(deRegCtx)
	}
}

func (w *Watchdog) runWithRegistry(cmd *exec.Cmd) {
	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, w.InterceptorSignals...)
	signal.Ignore(syscall.SIGURG, syscall.SIGWINCH)
	defer signal.Stop(signalChan)

	regCancel := w.asyncReg()
	once := sync.Once{}
	quitCtx, quitCancel := context.WithCancel(context.Background())
	defer quitCancel()
	for {
		select {
		case sig := <-signalChan:
			regCancel()
			w.OnEvent(sig)
			if sig != syscall.SIGCHLD && cmd != nil && cmd.Process != nil {
				if err := cmd.Process.Signal(sig); err == os.ErrProcessDone {
					once.Do(func() { w.syncDeReg() })
					return
				}
			}
			if sig == syscall.SIGCHLD {
				once.Do(func() { w.syncDeReg() })
				return
			} else {
				once.Do(func() { w.syncDeReg() })
				// 等待指定时间强制退出
				time.AfterFunc(w.QuitPostPone, quitCancel)
			}
		case <-quitCtx.Done():
			return
		}
	}
}

func (w *Watchdog) Spawn(name string, args ...string) error {
	w.normalize()
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	if w.Register != nil {
		w.runWithRegistry(cmd)
	}
	return cmd.Wait()
}
