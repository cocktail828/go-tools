package buffer

import (
	"errors"
	"sync/atomic"
)

var (
	ErrBufferFull  = errors.New("the ring buffer is full")
	ErrBufferEmpty = errors.New("the ring buffer is empty")
)

func high32(val int64) int64     { return val >> 32 }
func low32(val int64) int64      { return val & 0xFFFFFFFF }
func pack(high, low int64) int64 { return (high << 32) | (low & 0xFFFFFFFF) }

type RingBuffer struct {
	data []any
	size int64
	pos  atomic.Int64 // 高32位是head，低32位是tail
}

func NewRingBuffer(size int64) *RingBuffer {
	return &RingBuffer{
		data: make([]any, size),
		size: size,
	}
}

func (rb *RingBuffer) IsEmpty(val int64) bool { return high32(val) == low32(val) }
func (rb *RingBuffer) IsFull(val int64) bool  { return (low32(val)+1)%rb.size == high32(val) }

func (rb *RingBuffer) Enqueue(item any) error {
	for {
		pos := rb.pos.Load()
		head, tail := high32(pos), low32(pos)

		// check wether buffer is full
		nextTail := (tail + 1) % rb.size
		if nextTail == head {
			return ErrBufferFull
		}

		// try update with CAS
		newPos := pack(head, nextTail)
		if rb.pos.CompareAndSwap(pos, newPos) {
			rb.data[tail] = item
			return nil
		}
	}
}

func (rb *RingBuffer) Dequeue() (any, error) {
	for {
		pos := rb.pos.Load()
		head, tail := high32(pos), low32(pos)

		// check wether buffer is empty
		if head == tail {
			return nil, ErrBufferEmpty
		}

		// try update with CAS
		nextHead := (head + 1) % rb.size
		newPos := pack(nextHead, tail)
		if rb.pos.CompareAndSwap(pos, newPos) {
			item := rb.data[head]
			return item, nil
		}
	}
}
