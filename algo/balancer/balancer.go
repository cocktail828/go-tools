package balancer

type Node interface {
	Weight() int // 权重
	Value() any
}

type Healthy interface {
	Healthy() bool // 节点是否健康
}

type Balancer interface {
	Pick() Node
}
