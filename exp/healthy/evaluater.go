package healthy

import (
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
	*rolling.Rolling
}

// 适合仅有主动健康检测的场景
func NewCounterEvaluater(maxFailure, minSuccess int) Evaluater {
	return &counterEvaluater{
		MaxFailure: maxFailure,
		MinSuccess: minSuccess,
		healthy:    true,
		Rolling:    rolling.NewRolling(128),
	}
}

func (e *counterEvaluater) Check(err error) {
	if err == nil {
		e.DualIncrBy(1, 0)
	} else {
		e.DualIncrBy(0, 1)
	}
}

func (e *counterEvaluater) Alive() bool {
	posi, nega, _ := e.DualCount(39) // 获取过期 128ms*39=4.992s 的计数器信息
	if nega > int64(e.MaxFailure) {
		e.healthy = false
		return e.healthy
	}

	if posi > int64(e.MinSuccess) {
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
	*rolling.Rolling
}

// 基于成功率的健康状态检测
// 采用"滞后阈值"(Hysteresis Threshold), 可以有效避免系统在阈值附近频繁切换状态(抖动或振荡)
func NewPercentageEvaluater(minAlivePct, recoveryPct float32) Evaluater {
	return &percentageEvaluater{
		MinAlivePct: minAlivePct,
		RecoveryPct: recoveryPct,
		healthy:     true,
		Rolling:     rolling.NewRolling(128),
	}
}

func (e *percentageEvaluater) Check(err error) {
	if err == nil {
		e.DualIncrBy(1, 0)
	} else {
		e.DualIncrBy(0, 1)
	}
}

func (e *percentageEvaluater) Alive() bool {
	posi, nega, _ := e.DualCount(39)
	var pct float32
	if sum := posi + nega; sum > 0 {
		pct = float32(posi) / float32(sum)
	} else {
		return e.healthy // no ops was performed recently
	}

	if pct < e.MinAlivePct {
		e.healthy = false
		return e.healthy
	}

	if pct > e.RecoveryPct {
		e.healthy = true
	}
	return e.healthy
}
