package balancer

type Node interface {
	Weight() int // 权重
	Value() any
}

type Balancer interface {
	Pick() Node
}
