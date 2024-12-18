package event

type Event string

const (
	Success          Event = "success"
	Failure          Event = "failure"
	Reject           Event = "rejected"
	Canceled         Event = "canceled"
	ShortCircuit     Event = "short circuit"
	DeadlineExceeded Event = "deadline exceeded"
	FallBackSuccess  Event = "fallback success"
	FallbackFail     Event = "fallback failure"
)
