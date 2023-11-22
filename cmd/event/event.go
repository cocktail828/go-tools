package event

type Event int

const (
	None Event = iota
	Error
	Success
)
