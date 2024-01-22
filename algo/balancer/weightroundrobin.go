package balancer

import (
	"sync"

	"github.com/pkg/errors"
)

type weightRoundRobinBuilder struct{}

func (weightRoundRobinBuilder) Build() Balancer {
	return NewWeightRoundRobin()
}

type Weight interface {
	Weight() int
}

var _ Balancer = &weightRoundRobin{}

func init() {
	Register("weight-round-robin", weightRoundRobinBuilder{})
	Register("wrr", weightRoundRobinBuilder{})
}

type weightRoundRobin struct {
	array     []Weight
	mu        sync.Mutex
	busyArray []int
}

func NewWeightRoundRobin() *weightRoundRobin {
	return &weightRoundRobin{}
}

func (b *weightRoundRobin) Update(array []any) error {
	_array := make([]Weight, 0, len(array))
	for pos, mem := range array {
		if v, ok := mem.(Weight); !ok {
			return errors.Errorf("'Weight' interface not implemented, pos:%v", pos)
		} else {
			_array = append(_array, v)
		}
	}
	b.array = _array
	return nil
}

// nginx weighted round-robin balancing
// view: https://github.com/phusion/nginx/commit/27e94984486058d73157038f7950a0a36ecc6e35
func (b *weightRoundRobin) Pick() any {
	if len(b.array) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.busyArray) != len(b.array) {
		b.busyArray = make([]int, len(b.array))
	}

	allWeight := 0
	pos := -1
	for i := 0; i < len(b.array); i++ {
		c := b.array[i]
		if f, ok := c.(Validator); ok && !f.IsOK() {
			continue
		}

		allWeight += c.Weight()                             // 计算总权重
		b.busyArray[i] += c.Weight()                        // 当前权重加上权重
		if pos == -1 || b.busyArray[i] > b.busyArray[pos] { // 如果最优节点不存在或者当前节点由于最优节点，则赋值或者替换
			pos = i
		}
	}

	if pos != -1 {
		b.busyArray[pos] -= allWeight
	}
	return b.array[pos]
}
