package queue

import (
	"math"
	"sort"
	"sync"
	"testing"
)

// TestBasicOperations tests the fundamental functionality of PriorityQueue
func TestBasicOperationsPriorityQueue(t *testing.T) {
	// Create a new priority queue
	pq := New()

	// Test case: Enqueue elements with various priorities
	// Expected behavior: Elements should be dequeued in priority order (smallest first)
	elements := []float64{5, 3, 7, 8, 6, 2, 9}
	for _, e := range elements {
		pq.Push(e, e)
	}

	// Sort elements to verify correct order
	sort.Float64s(elements)
	for _, e := range elements {
		item, err := pq.Pop()
		if err != nil {
			t.Fatalf("Unexpected error while popping from queue: %v", err)
		}

		value, ok := item.(float64)
		if !ok {
			t.Fatalf("Popped item is not of expected type float64")
		}

		if e != value {
			t.Fatalf("Expected %v, got %v", e, value)
		}
	}
}

// TestPriorityUpdate tests the priority update functionality
func TestPriorityUpdate(t *testing.T) {
	pq := New()
	pq.Push("foo", 3)
	pq.Push("bar", 4)
	pq.UpdatePriority("bar", 2) // Update bar's priority to be higher than foo

	// After updating, bar should be dequeued first
	item, err := pq.Pop()
	if err != nil {
		t.Fatalf("Unexpected error while popping from queue: %v", err)
	}

	if item.(string) != "bar" {
		t.Fatal("Priority update failed: bar should be dequeued first")
	}
}

// TestQueueLength tests the Len method of PriorityQueue
func TestQueueLength(t *testing.T) {
	pq := New()
	// An empty queue should have length 0
	if pq.Len() != 0 {
		t.Fatal("Empty queue should have length of 0")
	}

	pq.Push("foo", 1)
	pq.Push("bar", 1)
	// After two inserts, queue length should be 2
	if pq.Len() != 2 {
		t.Fatal("Queue should have length of 2 after 2 inserts")
	}

	pq.Pop()
	// After one removal, queue length should be 1
	if pq.Len() != 1 {
		t.Fatal("Queue should have length of 1 after 1 removal")
	}
}

// TestDuplicateElements tests how the queue handles duplicate elements
func TestDuplicateElements(t *testing.T) {
	pq := New()
	pq.Push("foo", 2)
	pq.Push("bar", 3)
	pq.Push("bar", 1) // Attempt to push a duplicate element

	// Queue should ignore duplicate elements
	if pq.Len() != 2 {
		t.Fatal("Queue should ignore inserting the same element twice")
	}

	// Original element's priority should remain unchanged
	item, _ := pq.Pop()
	if item.(string) != "foo" {
		t.Fatal("Queue should ignore duplicate insert, not update existing item")
	}
}

// TestEmptyQueuePop tests popping from an empty queue
func TestEmptyQueuePop(t *testing.T) {
	pq := New()
	// Attempting to pop from an empty queue should return an error
	_, err := pq.Pop()
	if err == nil {
		t.Fatal("Should produce error when performing pop on empty queue")
	}
}

// TestNonExistingItemUpdate tests updating priority of a non-existing item
func TestNonExistingItemUpdate(t *testing.T) {
	pq := New()

	pq.Push("foo", 4)
	pq.UpdatePriority("bar", 5) // Attempt to update non-existent item

	// Queue length should remain unchanged
	if pq.Len() != 1 {
		t.Fatal("Update should not add items")
	}

	// Original item should still be present
	item, _ := pq.Pop()
	if item.(string) != "foo" {
		t.Fatalf("Update should not affect existing items, expected \"foo\", got \"%v\"", item.(string))
	}
}

// TestConcurrencySafety tests the priority queue under concurrent access
func TestConcurrencySafety(t *testing.T) {
	pq := New()
	var wg sync.WaitGroup
	const goroutines = 10
	const operations = 100

	// Concurrent pushes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				pq.Push(id*operations+j, float64(goroutines*operations-id*operations-j))
			}
		}(i)
	}
	wg.Wait()

	// Verify queue length
	expectedLen := goroutines * operations
	if pq.Len() != expectedLen {
		t.Fatalf("Expected queue length %d, got %d", expectedLen, pq.Len())
	}

	// Verify elements are dequeued in correct order
	prev := math.MaxInt64
	for i := 0; i < expectedLen; i++ {
		item, err := pq.Pop()
		if err != nil {
			t.Fatalf("Unexpected error during concurrent test: %v", err)
		}
		val := item.(int)
		if prev <= val {
			t.Fatalf("Elements out of order: expected %d <= %d", prev, val)
		}
		prev = val
	}
}

// TestDifferentElementTypes tests the queue with various hashable element types
func TestDifferentElementTypes(t *testing.T) {
	pq := New()

	// Push different types of hashable elements (avoiding slices and maps)
	pq.Push(42, 1.0)
	pq.Push("string", 2.0)
	pq.Push(3.14, 3.0)
	pq.Push(true, 4.0)
	pq.Push(struct{ name string }{
		name: "structValue",
	}, 5.0)

	// Verify correct order
	results := []any{42, "string", 3.14, true, struct{ name string }{
		name: "structValue",
	}}

	for i, expected := range results {
		item, err := pq.Pop()
		if err != nil {
			t.Fatalf("TestDifferentElementTypes: unexpected error at index %d: %v", i, err)
		}

		// Special handling for struct comparison since we can't directly compare struct instances
		if strct, ok := item.(struct{ name string }); ok {
			expectedStrct := expected.(struct{ name string })
			if strct.name != expectedStrct.name {
				t.Fatalf("TestDifferentElementTypes: struct name mismatch at index %d", i)
			}
		} else if item != expected {
			t.Fatalf("TestDifferentElementTypes: expected %v, got %v at index %d", expected, item, i)
		}
	}
}

// TestEdgeCasePriorities tests the queue with extreme priority values
func TestEdgeCasePriorities(t *testing.T) {
	pq := New()

	// Test with extreme priority values
	pq.Push("maxFloat", math.MaxFloat64)
	pq.Push("minFloat", math.SmallestNonzeroFloat64)
	pq.Push("negative", -1000.0)
	pq.Push("zero", 0.0)

	// Expected order: negative (-1000.0), zero (0.0), minFloat, maxFloat
	expectedOrder := []string{"negative", "zero", "minFloat", "maxFloat"}

	for i, expected := range expectedOrder {
		item, err := pq.Pop()
		if err != nil {
			t.Fatalf("TestEdgeCasePriorities: unexpected error at index %d: %v", i, err)
		}
		if item != expected {
			t.Fatalf("TestEdgeCasePriorities: expected %v, got %v at index %d", expected, item, i)
		}
	}
}

// TestMultipleUpdates tests multiple priority updates on the same item
func TestMultipleUpdates(t *testing.T) {
	pq := New()
	element := "dynamicItem"

	// Push with initial priority
	pq.Push(element, 100.0)

	// Multiple updates to test queue stability
	pq.UpdatePriority(element, 10.0)
	pq.UpdatePriority(element, 50.0)
	pq.UpdatePriority(element, 1.0)

	// After multiple updates, element should have the highest priority
	item, err := pq.Pop()
	if err != nil {
		t.Fatalf("TestMultipleUpdates: unexpected error: %v", err)
	}
	if item != element {
		t.Fatalf("TestMultipleUpdates: expected %v to be first after multiple updates", element)
	}
}
