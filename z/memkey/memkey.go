package memkey

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Option func(*Memkey)

func Interval(v time.Duration) Option {
	return func(m *Memkey) {
		m.interval = v
	}
}

func MaxItem(v int) Option {
	return func(m *Memkey) {
		m.maxitem = v
	}
}

func BatchSize(v int) Option {
	return func(m *Memkey) {
		m.batchSize = v
	}
}

func Percent(v float32) Option {
	return func(m *Memkey) {
		m.percent = v
	}
}

type wrapper struct {
	busy     atomic.Bool
	cb       func() bool
	accessAt time.Time
}

type Memkey struct {
	mu        sync.RWMutex
	interval  time.Duration
	maxitem   int
	batchSize int
	percent   float32
	cancel    context.CancelFunc
	eventChan chan time.Time
	busyMap   map[string]*wrapper
	idleMap   map[string]*wrapper
}

func New() *Memkey {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Memkey{
		cancel:    cancel,
		interval:  time.Millisecond * 100,
		maxitem:   1e5,
		batchSize: 20,
		percent:   0.25,
		eventChan: make(chan time.Time, 100),
		busyMap:   map[string]*wrapper{},
		idleMap:   map[string]*wrapper{},
	}
	go m.timedJob(ctx)
	return m
}

func (m *Memkey) Close() error {
	m.cancel()
	return nil
}

func (m *Memkey) timedJob(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-time.After(time.Millisecond * 100):
			m.publishEvent(now)
		case now := <-m.eventChan:
			if m.consumeAndEvict(now) {
				m.publishEvent(now)
			}
		}
	}
}

type event struct {
	keep bool
	now  time.Time
}

func (m *Memkey) publishEvent(now time.Time) {
	if len(m.eventChan) == 0 {
		m.eventChan <- now
	}
}

func (m *Memkey) consumeAndEvict(now time.Time) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loop := 0
	evicted := atomic.Int32{}
	for k, w := range m.busyMap {
		loop++
		if loop >= m.batchSize {
			break
		}

		delete(m.busyMap, k)
		if w.accessAt.Add(time.Second).After(now) || w.busy.Load() {
			m.idleMap[k] = w
			continue
		}

		ec := make(chan event, 2)
		go func() {
			w.busy.Store(true)
			defer w.busy.Store(false)

			select {
			case ec <- event{w.cb(), now}:
			case t := <-time.After(time.Millisecond * 50):
				ec <- event{true, t}
			}
		}()

		if e := <-ec; e.keep {
			w.accessAt = e.now
			m.idleMap[k] = w
		} else {
			evicted.Add(1)
		}
	}

	if len(m.busyMap) == 0 {
		m.busyMap, m.idleMap = m.idleMap, m.busyMap
	}
	return float32(evicted.Load())/float32(loop) >= m.percent
}

func (m *Memkey) Add(k string, f func() bool) bool {
	if k == "" || f == nil {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.busyMap)+len(m.idleMap) < m.maxitem {
		m.idleMap[k] = &wrapper{cb: f, accessAt: time.Now()}
		return true
	}
	return false
}

func (m *Memkey) Remove(k string) {
	if k == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.busyMap, k)
	delete(m.idleMap, k)
}
