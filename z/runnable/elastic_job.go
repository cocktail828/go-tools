package runnable

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/pkg/errors"
)

// Task represents a function that can be executed by the pool
type Task func()

var (
	ErrPoolClosed   = errors.New("elastic pool: already been closed")
	ErrPoolFull     = errors.New("elastic pool: task queue is full")
	ErrInvalidParam = errors.New("elastic pool: invalid parameters")
)

// Config contains configuration parameters for the worker pool
type Config struct {
	MaxWorkers      int           // Maximum number of workers
	MinWorkers      int           // Minimum number of workers
	PendingTaskNum  int           // Task queue length, default 1024
	ExpandThreshold float64       // Expand workers when task queue usage ratio exceeds this threshold, default 0.8
	ShrinkThreshold float64       // Shrink workers when task queue usage ratio falls below this threshold, default 0.3
	Period          time.Duration // Period is a round of time to check the task queue usage ratio, default 5 seconds
}

// Normalize validates and normalizes the configuration parameters
func (c *Config) Normalize() error {
	if c.MaxWorkers < c.MinWorkers || c.MinWorkers < 0 || c.MaxWorkers == 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters maxWorkers(%d), minWorkers(%d)", c.MaxWorkers, c.MinWorkers)
	}
	if c.PendingTaskNum <= 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters pendingTaskNum(%d)", c.PendingTaskNum)
	}
	if c.ShrinkThreshold <= 0 || c.ShrinkThreshold >= 1 {
		return errors.Wrapf(ErrInvalidParam, "check parameters shrinkThreshold(%f)", c.ShrinkThreshold)
	}
	if c.ExpandThreshold <= 0 || c.ExpandThreshold >= 1 {
		return errors.Wrapf(ErrInvalidParam, "check parameters expandThreshold(%f)", c.ExpandThreshold)
	}
	if c.Period <= 0 || c.Period > time.Minute {
		return errors.Wrapf(ErrInvalidParam, "check parameters peroid(%v), should be less than 1 minute", c.Period)
	}
	return nil
}

// DefaultConfig returns a default configuration for the worker pool
func DefaultConfig() Config {
	c := Config{
		MaxWorkers:     10,
		MinWorkers:     3,
		PendingTaskNum: 1024,
		Period:         5 * time.Second,
	}
	c.ExpandThreshold = 0.8
	c.ShrinkThreshold = 0.3
	return c
}

// ElasticJob implements a worker pool with elastic scaling capabilities
type ElasticJob struct {
	stateMux      sync.Mutex
	config        Config
	logger        xlog.Printer
	shrinkCh      chan struct{}      // Channel to signal workers to shrink
	runningCtx    context.Context    // Context for managing worker lifecycle
	runningCancel context.CancelFunc // Cancels the context
	wg            sync.WaitGroup     // Waits for all workers to finish
	mu            sync.RWMutex       // Mutex to protect access to the task channel
	closed        atomic.Bool        // Indicates whether the pool is closed
	taskCh        chan Task
}

// NewElasticJob creates a new worker pool with the given configuration
func NewElasticJob(c Config) (*ElasticJob, error) {
	if err := c.Normalize(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &ElasticJob{
		logger:        xlog.NoopPrinter{},
		taskCh:        make(chan Task, c.PendingTaskNum),
		shrinkCh:      make(chan struct{}, c.MaxWorkers),
		runningCtx:    ctx,
		runningCancel: cancel,
	}

	// Initialize with minimum workers
	for i := 0; i < c.MinWorkers; i++ {
		p.spawn()
	}

	// Start the elastic scaling goroutine
	p.wg.Add(1)
	go p.elastic()

	return p, nil
}

func (p *ElasticJob) SetLogger(logger xlog.Printer) {
	p.logger = logger
}

// Tune updates the configuration of the worker pool
// However, the 'PendingTaskNum' parameter will not be affect
func (p *ElasticJob) Tune(c Config) error {
	if err := c.Normalize(); err != nil {
		return err
	}
	p.stateMux.Lock()
	defer p.stateMux.Unlock()
	p.config = c
	return nil
}

// elastic manages the worker pool scaling based on task queue usage
func (p *ElasticJob) elastic() {
	defer p.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var highWaterTime time.Time           // Last time high water threshold was reached
	var lowWaterTime time.Time            // Last time low water threshold was reached
	currentWorkers := p.config.MinWorkers // already spawned at start

	for {
		select {
		case <-ticker.C:
			func() {
				// concurrency safe
				if !p.stateMux.TryLock() {
					return
				}
				defer p.stateMux.Unlock()

				currentTime := time.Now()
				taskQueueLen := float64(len(p.taskCh))
				taskQueueCap := float64(cap(p.taskCh))
				currentUsage := taskQueueLen / taskQueueCap
				step := max((p.config.MaxWorkers-p.config.MinWorkers)/10, 1)

				// 1. Check if we need to expand workers
				if currentUsage >= p.config.ExpandThreshold && currentWorkers < p.config.MaxWorkers {
					if highWaterTime.IsZero() {
						highWaterTime = currentTime
					} else if currentTime.Sub(highWaterTime) >= p.config.Period {
						for i := 0; i < step && currentWorkers < p.config.MaxWorkers; i++ {
							p.spawn()
							currentWorkers++
						}
						p.logger.Printf("expand workers to %d", currentWorkers)
						highWaterTime = time.Time{}
					}
				} else {
					highWaterTime = time.Time{}
				}

				// 2. Check if we need to shrink workers
				if currentUsage < p.config.ShrinkThreshold && currentWorkers > p.config.MinWorkers {
					if lowWaterTime.IsZero() {
						lowWaterTime = currentTime
					} else if currentTime.Sub(lowWaterTime) >= p.config.Period {
						for i := 0; i < step && currentWorkers > p.config.MinWorkers; i++ {
							select {
							case p.shrinkCh <- struct{}{}:
								currentWorkers--
							default:
							}
						}
						p.logger.Printf("shrink workers to %d", currentWorkers)
						lowWaterTime = time.Time{}
					}
				} else {
					lowWaterTime = time.Time{}
				}
			}()

		case <-p.runningCtx.Done():
			return
		}
	}
}

// spawn creates a new worker that processes tasks
func (p *ElasticJob) spawn() {
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()

		for {
			select {
			case <-p.runningCtx.Done():
				return
			case <-p.shrinkCh: // Elastic worker stopped due to idleness
				return
			case task, ok := <-p.taskCh:
				if !ok {
					return
				}
				p.execute(task)
			default:
				return
			}
		}
	}()
}

type CloseResult struct {
	ch  <-chan Task
	err error
}

// Is returns true if the underlying error is the same as the target error
func (cr *CloseResult) Is(target error) bool { return errors.Is(cr.err, target) }

func (cr *CloseResult) Error() string { return cr.err.Error() }

// Chan returns the task channel of the worker pool
// Users can drain the task channel after closing the pool but it returns a error
func (cr *CloseResult) Chan() <-chan Task { return cr.ch }

// Close shutdown the worker pool, drain the task channel and wait for all tasks to complete
// It returns an error if the context deadline is exceeded
func (p *ElasticJob) Close(ctx context.Context) *CloseResult {
	p.runningCancel()

	// concurrency safe when closing task channel
	if p.mu.TryLock() {
		// ensure that the task channel is closed once
		if p.closed.CompareAndSwap(false, true) {
			close(p.taskCh)
		}
		p.mu.Unlock()
	}

loop:
	for {
		select {
		case task, ok := <-p.taskCh:
			// drain the channel
			if !ok {
				break loop
			}
			p.execute(task)

		case <-ctx.Done():
			p.wg.Wait()
			return &CloseResult{p.taskCh, ctx.Err()}
		}
	}
	p.wg.Wait()

	return nil
}

func (p *ElasticJob) execute(task Task) {
	defer func() {
		if err := recover(); err != nil {
			p.logger.Printf("Task panic: %v\n", err)
		}
	}()
	task()
}

// Submit adds a task to the worker pool
func (p *ElasticJob) Submit(task Task) (err error) {
	// in case of write to a closed channel
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case p.taskCh <- task:
		return nil
	default:
		return ErrPoolFull
	}
}
