package balancer

type wrrBalancer struct {
	nodeArray
	busyArray []int
}

func NewWRR(nodes []Node) Balancer {
	return &wrrBalancer{
		nodeArray: nodeArray{nodes: nodes},
		busyArray: make([]int, len(nodes)),
	}
}

func (b *wrrBalancer) Update(nodes []Node) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.updateLocked(nodes)
	b.busyArray = make([]int, len(nodes))
}

// nginx weighted round-robin balancing
// view: https://github.com/phusion/nginx/commit/27e94984486058d73157038f7950a0a36ecc6e35
func (b *wrrBalancer) Pick() Node {
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	allWeight := 0
	pos := -1
	for i := 0; i < len(b.nodes); i++ {
		c := b.nodes[i]
		allWeight += c.Weight()                             // 计算总权重
		b.busyArray[i] += c.Weight()                        // 当前权重加上权重
		if pos == -1 || b.busyArray[i] > b.busyArray[pos] { // 如果最优节点不存在或者当前节点由于最优节点，则赋值或者替换
			pos = i
		}
	}

	if pos != -1 {
		b.busyArray[pos] -= allWeight
		return WrapNode{Node: b.nodes[pos], nodeArrayRemove: b}
	}

	return nil
}
