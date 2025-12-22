package hashring

import (
	"math"
	"strconv"
	"sync"
	"testing"

	"github.com/cocktail828/go-tools/algo/hash/murmur3"
)

const (
	// Test nodes
	node1 = "192.168.1.1"
	node2 = "192.168.1.2"
	node3 = "192.168.1.3"
	node4 = "192.168.1.4"
)

// TestHashRingCreation tests the creation of a new hash ring with default and custom options
func TestHashRingCreation(t *testing.T) {
	// Test with default options
	hr := New()
	if hr == nil {
		t.Fatal("Expected non-nil HashRing")
	}
	if hr.virualSpots != DefaultVirualSpots {
		t.Fatalf("Expected virualSpots to be %d, got %d", DefaultVirualSpots, hr.virualSpots)
	}

	// Test with custom virtual spots
	customSpots := 100
	hr = New(WithVirtualSpots(customSpots))
	if hr.virualSpots != customSpots {
		t.Fatalf("Expected virualSpots to be %d, got %d", customSpots, hr.virualSpots)
	}

	// Test with custom hash function
	customHash := func(s string) uint32 {
		return murmur3.Sum32([]byte(s))
	}
	hr = New(WithHash(customHash))
	if hr.hashFunc == nil {
		t.Fatal("Expected non-nil hash function")
	}
}

// TestAddAndGet tests adding nodes and getting the responsible node for a key
func TestAddAndGet(t *testing.T) {
	hr := New()

	// Add a single node
	hr.Add(node1, 1)

	// Verify that we can get the node
	key := "test-key"
	node := hr.Get(key)
	if node != node1 {
		t.Fatalf("Expected node %s, got %s", node1, node)
	}

	// Add multiple nodes
	hr.Add(node2, 1)
	hr.Add(node3, 1)

	// Verify that all nodes are in the hash ring
	nodesMap := make(map[string]bool)
	nodesMap[node1] = false
	nodesMap[node2] = false
	nodesMap[node3] = false

	// Check multiple keys to ensure distribution
	for i := 0; i < 100; i++ {
		key := "key-" + strconv.Itoa(i)
		node := hr.Get(key)
		if node == "" {
			t.Fatalf("Expected non-empty node for key %s", key)
		}
		nodesMap[node] = true
	}

	// Ensure all nodes were used
	for node, used := range nodesMap {
		if !used {
			t.Fatalf("Node %s was never selected", node)
		}
	}
}

// TestAddMany tests adding multiple nodes with weights at once
func TestAddMany(t *testing.T) {
	hr := New()

	// Add multiple nodes with weights
	nodes := map[string]int{
		node1: 1,
		node2: 1,
		node3: 1,
	}
	hr.AddMany(nodes)

	// Verify nodes are added
	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := "key-" + strconv.Itoa(i)
		node := hr.Get(key)
		counts[node]++
	}

	// Ensure all nodes are present
	for node := range nodes {
		if counts[node] == 0 {
			t.Fatalf("Node %s was never selected", node)
		}
	}
}

// TestRemove tests removing a node from the hash ring
func TestRemove(t *testing.T) {
	hr := New()
	hr.Add(node1, 1)
	hr.Add(node2, 1)

	// Verify node1 is present
	key := "test-key"
	// nodeBefore := hr.Get(key)

	// Remove node1
	hr.Remove(node1)

	// Verify node1 is no longer present
	if hr.Get(key) == node1 {
		t.Fatalf("Node %s should have been removed", node1)
	}

	// After removing all nodes, Get should return empty string
	hr.Remove(node2)
	if node := hr.Get(key); node != "" {
		t.Fatalf("Expected empty string for empty hash ring, got %s", node)
	}
}

// TestUpdate tests updating a node's weight
func TestUpdate(t *testing.T) {
	hr := New()
	hr.Add(node1, 1)
	hr.Add(node2, 1)

	// Get initial distribution
	initialCounts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := "key-" + strconv.Itoa(i)
		node := hr.Get(key)
		initialCounts[node]++
	}

	// Add node2's weight to be higher
	hr.Add(node2, 5)

	// Get new distribution
	updatedCounts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := "key-" + strconv.Itoa(i)
		node := hr.Get(key)
		updatedCounts[node]++
	}

	// Verify node2 has more keys after weight increase
	if updatedCounts[node2] <= initialCounts[node2] {
		t.Fatalf("Node2 should have more keys after weight increase")
	}
}

// TestWeights tests that node weights affect distribution correctly
func TestWeights(t *testing.T) {
	hr := New()

	// Add nodes with different weights
	hr.Add(node1, 1) // 1x weight
	hr.Add(node2, 2) // 2x weight
	hr.Add(node3, 3) // 3x weight

	// Test distribution
	counts := make(map[string]int)
	samples := 10000
	for i := 0; i < samples; i++ {
		key := "key-" + strconv.Itoa(i)
		node := hr.Get(key)
		counts[node]++
	}

	// Calculate percentages
	totalWeight := 1 + 2 + 3
	expectedNode1 := float64(counts[node1]) / float64(samples) * float64(totalWeight)
	expectedNode2 := float64(counts[node2]) / float64(samples) * float64(totalWeight)
	expectedNode3 := float64(counts[node3]) / float64(samples) * float64(totalWeight)

	// Allow for some distribution variance
	tolerance := 0.2
	if !approximatelyEqual(expectedNode1, 1.0, tolerance) {
		t.Fatalf("Node1 distribution (%.2f) not close to expected weight (1.0)", expectedNode1)
	}
	if !approximatelyEqual(expectedNode2, 2.0, tolerance) {
		t.Fatalf("Node2 distribution (%.2f) not close to expected weight (2.0)", expectedNode2)
	}
	if !approximatelyEqual(expectedNode3, 3.0, tolerance) {
		t.Fatalf("Node3 distribution (%.2f) not close to expected weight (3.0)", expectedNode3)
	}
}

// TestVirtualSpots tests that different virtual spot counts affect distribution granularity
func TestVirtualSpots(t *testing.T) {
	// Test with low virtual spots
	lowVirtualSpots := 10
	hrLow := New(WithVirtualSpots(lowVirtualSpots))
	hrLow.AddMany(map[string]int{node1: 1, node2: 1, node3: 1})

	// Test with high virtual spots
	highVirtualSpots := 1000
	hrHigh := New(WithVirtualSpots(highVirtualSpots))
	hrHigh.AddMany(map[string]int{node1: 1, node2: 1, node3: 1})

	// Calculate distribution for both
	samples := 10000
	lowCounts := make(map[string]int)
	highCounts := make(map[string]int)

	for i := 0; i < samples; i++ {
		key := "key-" + strconv.Itoa(i)
		lowCounts[hrLow.Get(key)]++
		highCounts[hrHigh.Get(key)]++
	}

	// The high virtual spots should have a more even distribution
	lowDeviation := calculateDeviation(lowCounts, samples, 3)
	highDeviation := calculateDeviation(highCounts, samples, 3)

	if highDeviation >= lowDeviation {
		t.Fatalf("Expected lower distribution deviation with high virtual spots")
	}
}

// TestConsistency tests that adding/removing nodes doesn't disrupt existing key mapping unnecessarily
func TestConsistency(t *testing.T) {
	// Create initial hash ring with three nodes
	hr := New()
	hr.AddMany(map[string]int{node1: 1, node2: 1, node3: 1})

	// Get initial mapping for some keys
	keyNodeMap := make(map[string]string)
	keys := 1000
	for i := 0; i < keys; i++ {
		key := "persistent-key-" + strconv.Itoa(i)
		keyNodeMap[key] = hr.Get(key)
	}

	// Add a new node
	hr.Add(node4, 1)

	// Calculate how many keys changed their node assignment
	changed := 0
	for key, originalNode := range keyNodeMap {
		newNode := hr.Get(key)
		if newNode != originalNode {
			changed++
		}
	}

	// Check that most keys remain mapped to the same node (consistent hashing property)
	changeRatio := float64(changed) / float64(keys)
	expectedMaxChangeRatio := 0.25 // Approximately 1/(N+1) where N is the original number of nodes
	if changeRatio > expectedMaxChangeRatio {
		t.Fatalf("Too many keys changed nodes: %.2f%% (expected <= %.2f%%)",
			changeRatio*100, expectedMaxChangeRatio*100)
	}
}

// TestConcurrencySafety is a simple concurrency test to check for race conditions
func TestConcurrencySafety(t *testing.T) {
	// This is a basic test to check for race conditions
	// When running with -race flag, it will detect any data races

	hr := New()
	hr.AddMany(map[string]int{node1: 1, node2: 1, node3: 1})

	// Run concurrent operations
	var wg sync.WaitGroup
	operations := 1000

	// Concurrent gets
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "concurrent-key-" + strconv.Itoa(i)
			hr.Get(key)
		}(i)
	}

	// Concurrent updates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Periodically update weights
			weight := (i % 5) + 1
			hr.Add(node1, weight)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

// Helper functions
func approximatelyEqual(a, b, tolerance float64) bool {
	diff := math.Abs(a - b)
	return diff <= tolerance
}

func calculateDeviation(counts map[string]int, total int, nodeCount int) float64 {
	// Calculate standard deviation of distribution
	expected := float64(total) / float64(nodeCount)
	variance := 0.0

	for _, count := range counts {
		diff := float64(count) - expected
		variance += diff * diff
	}

	variance /= float64(nodeCount)
	return math.Sqrt(variance) / expected // Normalized standard deviation
}
