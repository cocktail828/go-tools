package balancer

type Validator interface {
	IsOK() bool
}

type BalancerBuilder interface {
	Build() Balancer
}

type Balancer interface {
	Pick() any
	Update([]any) error
}

var (
	m = map[string]BalancerBuilder{}
)

func Register(n string, b BalancerBuilder) {
	m[n] = b
}

func FindBalancer(name string) BalancerBuilder {
	return m[name]
}
