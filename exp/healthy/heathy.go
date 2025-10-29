package healthy

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/xlog"
)

type Keepalive interface {
	Alive() bool                                      // 当前健康状态
	Check(err error)                                  // 记录错误
	Background(itvl time.Duration) context.CancelFunc // 启动后台探活
}

type timedVal struct {
	val bool
	tm  time.Time
}

type keepaliveImpl struct {
	Evaluater // 健康状态评估
	Liveness  // 健康检查器
	logger    xlog.Printer
	running   atomic.Bool
	healthy   timedVal
	mu        sync.Mutex
}

func NewKeepalive(el Evaluater, lv Liveness, logger xlog.Printer) Keepalive {
	return &keepaliveImpl{
		Evaluater: el,
		Liveness:  lv,
		logger:    logger,
	}
}

func (ka *keepaliveImpl) Alive() bool {
	tv := ka.healthy

	if now := time.Now(); now.Sub(tv.tm) > time.Millisecond*100 {
		if ka.mu.TryLock() {
			defer ka.mu.Unlock()
			ka.healthy = timedVal{ka.Evaluater.Alive(), now}
		}
	}
	return tv.val
}

// 被动探活, 业务主动反馈每次执行的情况
func (ka *keepaliveImpl) Check(err error) {
	ka.Evaluater.Check(err)
}

// 主动后台探活
func (ka *keepaliveImpl) Background(itvl time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	if itvl == 0 {
		return cancel
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				ka.logger.Printf("keepaliveImpl Background goroutine panic: %v", r)
			}
		}()

		// already spawn a background goroutine
		if !ka.running.CompareAndSwap(false, true) {
			return
		}
		defer ka.running.Store(false)

		ticker := time.NewTicker(itvl)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ka.Check(ka.Probe())
			}
		}
	}()
	return cancel
}
