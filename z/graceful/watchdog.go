package graceful

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Watchdog struct {
	Postpone   time.Duration
	Register   func()
	DeRegister func()
}

func (w Watchdog) Respawn(c <-chan os.Signal, name string, args ...string) error {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	timer := time.NewTimer(w.Postpone)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done(): // child exit?
				return

			case <-timer.C:
				// check liveless
				if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
					return
				}
				w.Register()

			case sig := <-c:
				// start fail?
				if cmd.Process == nil {
					return
				}
				if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
					return
				}

				switch sig {
				case syscall.SIGCHLD:
					// child exit?
					return
				default:
					// proxy pass signal
					cmd.Process.Signal(sig)
				}
			}
		}
	}()

	err := cmd.Run()
	cancel()
	w.DeRegister()
	return err
}
