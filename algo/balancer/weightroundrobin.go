package balancer

import (
	"sync"

	"github.com/cocktail828/go-tools/z"
)

type wrrBalancer struct {
	sync.Mutex
	array     []Node
	busyArray []int
}

func NewWRR() Balancer {
	return &wrrBalancer{}
}

func (b *wrrBalancer) Update(array []Node) {
	z.WithLock(b, func() {
		b.array = array
		b.busyArray = make([]int, len(b.array))
	})
}

// nginx weighted round-robin balancing
// view: https://github.com/phusion/nginx/commit/27e94984486058d73157038f7950a0a36ecc6e35
func (b *wrrBalancer) Pick() (n Node) {
	if len(b.array) == 0 {
		return
	}

	z.WithLock(b, func() {
		allWeight := 0
		pos := -1
		for i := 0; i < len(b.array); i++ {
			c := b.array[i]
			allWeight += c.Weight()                             // 计算总权重
			b.busyArray[i] += c.Weight()                        // 当前权重加上权重
			if pos == -1 || b.busyArray[i] > b.busyArray[pos] { // 如果最优节点不存在或者当前节点由于最优节点，则赋值或者替换
				pos = i
			}
		}

		if pos != -1 {
			b.busyArray[pos] -= allWeight
			n = b.array[pos]
		}
	})
	return
}
