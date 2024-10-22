package timewheel_test

import (
	"sync"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/timewheel"
)

type MockTask struct {
	handler func()
}

func (m *MockTask) Handle() {
	if m.handler != nil {
		m.handler()
	}
}

func TestTimeWheel_AddAndRemoveTask(t *testing.T) {
	tw := timewheel.New(10*time.Millisecond, 5)
	if tw == nil {
		t.Fatal("Expected TimeWheel to be created successfully")
	}

	wg := sync.WaitGroup{}

	taskAdded := false
	mockTask := &MockTask{
		handler: func() {
			taskAdded = true
			wg.Done()
		},
	}

	wg.Add(1)
	tw.Add(mockTask, time.Millisecond*20)

	tw.Start()
	defer tw.Stop()

	time.Sleep(30 * time.Millisecond)

	wg.Wait()
	if !taskAdded {
		t.Error("Expected task to be executed, but it was not")
	}

	tw.Remove(mockTask)

	taskExecuted := false
	mockTask2 := &MockTask{
		handler: func() {
			taskExecuted = true
			wg.Done()
		},
	}
	wg.Add(1)
	tw.Add(mockTask2, time.Millisecond*20)

	time.Sleep(30 * time.Millisecond)

	wg.Wait()
	if !taskExecuted {
		t.Error("Expected task to be executed after re-adding, but it was not")
	}
}

func TestTimeWheel_Concurrency(t *testing.T) {
	tw := timewheel.New(10*time.Millisecond, 5)
	if tw == nil {
		t.Fatal("Expected TimeWheel to be created successfully")
	}

	var wg sync.WaitGroup
	const numTasks = 10
	taskExecuted := make([]bool, numTasks)

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		taskID := i
		tw.Add(&MockTask{
			handler: func() {
				taskExecuted[taskID] = true
				wg.Done()
			},
		}, time.Millisecond*20)
	}

	tw.Start()
	defer tw.Stop()

	time.Sleep(30 * time.Millisecond)
	wg.Wait()
	for i := 0; i < numTasks; i++ {
		if !taskExecuted[i] {
			t.Errorf("Expected task %d to be executed, but it was not", i)
		}
	}
}
