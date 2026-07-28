package healthy

import (
	"sync/atomic"

	"github.com/cocktail828/go-tools/algo/rolling"
	"github.com/cocktail828/go-tools/z/timex"
)

type Evaluater interface {
	Check(error) // 记录成功或失败
	Alive() bool // 返回当前健康状态
}

// 基于计数的健康状态评估
type counterEvaluater struct {
	MaxFailure int // 超过此值将转为亚健康状态
	MinSuccess int // 超过此值将恢复健康状态

	// privates
	healthy atomic.Bool // 健康状态 (thread-safe)
	negC    *rolling.SlidingWindow
	posC    *rolling.SlidingWindow
}

// 适合仅有主动健康检测的场景
func NewCounterEvaluater(maxFailure, minSuccess int) Evaluater {
	e := &counterEvaluater{
		MaxFailure: maxFailure,
		MinSuccess: minSuccess,
		negC:       rolling.NewSlidingWindow(128),
		posC:       rolling.NewSlidingWindow(128),
	}
	e.healthy.Store(true)
	return e
}

func (e *counterEvaluater) Check(err error) {
	if err == nil {
		e.posC.IncrBy(1)
	} else {
		e.negC.IncrBy(1)
	}
}

func (e *counterEvaluater) Alive() bool {
	nsec := timex.UnixNano()
	nega, _ := e.negC.At(nsec).Estimate(24) // 获取过期 128ms*24=3.072s 的计数器信息
	posi, _ := e.posC.At(nsec).Estimate(24) // 获取过期 128ms*24=3.072s 的计数器信息
	if nega > int64(e.MaxFailure) {
		e.healthy.Store(false)
		return false
	}

	if posi > int64(e.MinSuccess) {
		e.healthy.Store(true)
	}

	return e.healthy.Load()
}

// 基于百分比的健康状态评估
type percentageEvaluater struct {
	MinAlivePct float32 // 最小健康水位
	RecoveryPct float32 // 恢复健康的阈值

	// privates
	healthy atomic.Bool // 健康状态 (thread-safe)
	negC    *rolling.SlidingWindow
	posC    *rolling.SlidingWindow
}

// 基于成功率的健康状态检测
// 采用"滞后阈值"(Hysteresis Threshold), 可以有效避免系统在阈值附近频繁切换状态(抖动或振荡)
func NewPercentageEvaluater(minAlivePct, recoveryPct float32) Evaluater {
	e := &percentageEvaluater{
		MinAlivePct: minAlivePct,
		RecoveryPct: recoveryPct,
		negC:        rolling.NewSlidingWindow(128),
		posC:        rolling.NewSlidingWindow(128),
	}
	e.healthy.Store(true)
	return e
}

func (e *percentageEvaluater) Check(err error) {
	if err == nil {
		e.posC.IncrBy(1)
	} else {
		e.negC.IncrBy(1)
	}
}

func (e *percentageEvaluater) Alive() bool {
	nsec := timex.UnixNano()
	nega, _ := e.negC.At(nsec).Estimate(24) // 获取过期 128ms*24=3.072s 的计数器信息
	posi, _ := e.posC.At(nsec).Estimate(24) // 获取过期 128ms*24=3.072s 的计数器信息
	var pct float32
	if sum := posi + nega; sum > 0 {
		pct = float32(posi) / float32(sum)
	} else {
		return e.healthy.Load() // no ops was performed recently
	}

	if pct < e.MinAlivePct {
		e.healthy.Store(false)
		return false
	}

	if pct > e.RecoveryPct {
		e.healthy.Store(true)
	}
	return e.healthy.Load()
}
