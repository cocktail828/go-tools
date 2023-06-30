package loadbalance

import (
	"sync"
)

type WeightCollection interface {
	// Validate will varify validity of element.
	Validate(int) bool
	// Len is the number of elements in the collection.
	Len() int
	// Weight is the weight of element.
	Weight(int) int
}

type WeightRoundRobin struct {
	mu        sync.Mutex
	len       int
	curWeight []int
}

func NewWeightRoundRobin() *WeightRoundRobin {
	return &WeightRoundRobin{}
}

// nginx weighted round-robin balancing
// view: https://github.com/phusion/nginx/commit/27e94984486058d73157038f7950a0a36ecc6e35
func (lhs *WeightRoundRobin) Get(c WeightCollection) int {
	lhs.mu.Lock()
	defer lhs.mu.Unlock()

	if lhs.len != c.Len() {
		lhs.len = c.Len()
		lhs.curWeight = make([]int, c.Len())
	}

	allWeight := 0
	pos := -1
	for i := 0; i < c.Len(); i++ {
		if !c.Validate(i) {
			continue
		}

		allWeight += c.Weight(i)                                // 计算总权重
		lhs.curWeight[i] += c.Weight(i)                         // 当前权重加上权重
		if pos == -1 || lhs.curWeight[i] > lhs.curWeight[pos] { // 如果最优节点不存在或者当前节点由于最优节点，则赋值或者替换
			pos = i
		}
	}

	if pos != -1 {
		lhs.curWeight[pos] -= allWeight
	}
	return pos
}
