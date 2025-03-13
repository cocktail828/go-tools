package healthy

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
)

type Evaluater interface {
	Check(error) // 记录成功或失败
	Alive() bool // 返回当前健康状态, 不允许并发调用
}

// 基于计数的健康状态评估
type counterEvaluater struct {
	MaxFailure int // 超过此值将转为亚健康状态
	MinSuccess int // 超过此值将恢复健康状态

	// privates
	healthy bool // 健康状态
	*rolling.DualRolling
}

// 适合仅有主动健康检测的场景
func NewCounterEvaluater(maxFailure, minSuccess int) Evaluater {
	return &counterEvaluater{
		MaxFailure:  maxFailure,
		MinSuccess:  minSuccess,
		healthy:     true,
		DualRolling: rolling.NewDualRolling(128, 256),
	}
}

func (e *counterEvaluater) Check(err error) {
	if err == nil {
		e.Incr(1, 0)
	} else {
		e.Incr(0, 1)
	}
}

func (e *counterEvaluater) Alive() bool {
	posi, nega, _ := e.Count(50) // 获取过期 128ms*50=6.4s 的计数器信息
	if e.healthy && nega > e.MaxFailure {
		e.healthy = false
	}

	if !e.healthy && posi > e.MinSuccess {
		e.healthy = true
	}

	return e.healthy
}

// 基于百分比的健康状态评估
type percentageEvaluater struct {
	MinAlivePct float32 // 最小健康水位
	RecoveryPct float32 // 恢复健康的阈值

	// privates
	healthy bool // 健康状态
	*rolling.DualRolling
}

// 基于成功率的健康状态检测
// 采用"滞后阈值"(Hysteresis Threshold), 可以有效避免系统在阈值附近频繁切换状态(抖动或振荡)
func NewPercentageEvaluater(minAlivePct, recoveryPct float32) Evaluater {
	return &percentageEvaluater{
		MinAlivePct: minAlivePct,
		RecoveryPct: recoveryPct,
		healthy:     true,
		DualRolling: rolling.NewDualRolling(128, 256),
	}
}

func (e *percentageEvaluater) Check(err error) {
	if err == nil {
		e.Incr(1, 0)
	} else {
		e.Incr(0, 1)
	}
}

func (e *percentageEvaluater) Alive() bool {
	posi, nega, _ := e.Count(50)
	var pct float32
	if sum := posi + nega; sum > 0 {
		pct = float32(posi) / float32(posi+nega)
	} else {
		// no ops was performed recently
		return e.healthy
	}

	if pct < e.MinAlivePct {
		e.healthy = false
	} else if pct > e.RecoveryPct {
		e.healthy = true
	}
	return e.healthy
}

type Keepalive interface {
	Alive() bool                                      // 当前健康状态
	Check(err error)                                  // 记录错误
	Background(itvl time.Duration) context.CancelFunc // 启动后台探活
}

type ttled struct {
	val bool
	tm  time.Time
}

type keepaliveImpl struct {
	Evaluater // 健康状态评估
	Liveness  // 健康检查器
	running   atomic.Bool
	healthy   ttled
	mu        sync.Mutex
}

func NewKeepalive(el Evaluater, lv Liveness) Keepalive {
	return &keepaliveImpl{
		Evaluater: el,
		Liveness:  lv,
	}
}

func (ka *keepaliveImpl) Alive() bool {
	tv := ka.healthy

	now := time.Now()
	if now.Sub(tv.tm) > time.Millisecond*100 {
		if ka.mu.TryLock() {
			defer ka.mu.Unlock()
			tv = ttled{ka.Evaluater.Alive(), now}
			ka.healthy = tv
		}
	}
	return tv.val
}

func (ka *keepaliveImpl) checkInternal(err error, active bool) {
	ka.Evaluater.Check(err)
}

// 被动探活, 业务主动反馈每次执行的情况
func (ka *keepaliveImpl) Check(err error) {
	ka.checkInternal(err, false)
}

// 主动后台探活
func (ka *keepaliveImpl) Background(itvl time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("keepaliveImpl Background goroutine panic: %v", r)
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
				ka.checkInternal(ka.Probe(), true)
			}
		}
	}()
	return cancel
}
