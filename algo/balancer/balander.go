package balancer

type Validator interface {
	IsOK() bool
}

type Balancer interface {
	Pick() Validator
	Update([]Validator)
}
