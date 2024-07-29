package nacs

import "errors"

var (
	ErrNotImpl = errors.New("the method is not implement")
)

type Event string

const (
	ADD Event = "add"
	DEL Event = "del"
	CHG Event = "chg"
)
