package timewheel

import (
	"sync"
	"time"
)

type Task interface {
	Handle()
}

type slot struct {
	id    int
	tasks map[Task]struct{}
}

func (s *slot) add(t Task) {
	s.tasks[t] = struct{}{}
}

func (s *slot) remove(t Task) {
	delete(s.tasks, t)
}

type TimeWheel struct {
	tickDuration     time.Duration  // 每个槽的时间间隔
	ticksPerWheel    int            // 槽的数量
	currentTickIndex int            // 当前槽的索引
	ticker           *time.Ticker   // 定时器
	mu               sync.RWMutex   // 读写锁
	wheels           []*slot        // 时间轮的槽
	indicator        map[Task]*slot // 指向每个任务所在槽的指针
	wg               *sync.WaitGroup

	quitChan chan struct{}
}

func New(tickDuration time.Duration, ticksPerWheel int) *TimeWheel {
	if tickDuration < 1 || ticksPerWheel < 1 {
		return nil
	}

	tw := &TimeWheel{
		tickDuration:     tickDuration,
		ticksPerWheel:    ticksPerWheel,
		currentTickIndex: 0,
		quitChan:         make(chan struct{}),
		indicator:        make(map[Task]*slot),
		wheels:           make([]*slot, ticksPerWheel),
		wg:               &sync.WaitGroup{},
	}

	for i := 0; i < ticksPerWheel; i++ {
		tw.wheels[i] = &slot{
			id:    i,
			tasks: make(map[Task]struct{}),
		}
	}

	return tw
}

func (tw *TimeWheel) Add(t Task, dur time.Duration) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.removeLocked(t)

	idx := (tw.getCurrentTickIndex() + int(dur/tw.tickDuration)) % tw.ticksPerWheel
	slot := tw.wheels[idx]
	slot.add(t)
	tw.indicator[t] = slot
}

func (tw *TimeWheel) Remove(t Task) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.removeLocked(t)
}

func (tw *TimeWheel) removeLocked(t Task) {
	if v, ok := tw.indicator[t]; ok {
		v.remove(t)
		delete(tw.indicator, t)
	}
}

func (tw *TimeWheel) getCurrentTickIndex() int {
	return tw.currentTickIndex
}

func (tw *TimeWheel) Stop() {
	close(tw.quitChan)
	tw.wg.Wait()
}

func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.tickDuration)

	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		for {
			select {
			case <-tw.quitChan:
				tw.ticker.Stop()
				return

			case <-tw.ticker.C:
				tasks := []Task{}
				tw.mu.Lock()
				slot := tw.wheels[tw.currentTickIndex]
				for t := range slot.tasks {
					slot.remove(t)
					delete(tw.indicator, t)
					tasks = append(tasks, t)
				}
				tw.mu.Unlock()

				tw.currentTickIndex = (tw.currentTickIndex + 1) % tw.ticksPerWheel

				tw.wg.Add(1)
				go func() {
					defer tw.wg.Done()
					for _, t := range tasks {
						t.Handle()
					}
				}()
			}
		}
	}()
}
