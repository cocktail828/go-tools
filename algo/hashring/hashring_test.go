package hashring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	node1 = "192.168.1.1"
	node2 = "192.168.1.2"
	node3 = "192.168.1.3"
)

func getNodesCount(nodes nodesArray) (int, int, int) {
	node1Count := 0
	node2Count := 0
	node3Count := 0

	for _, node := range nodes {
		if node.nodeKey == node1 {
			node1Count += 1
		}
		if node.nodeKey == node2 {
			node2Count += 1

		}
		if node.nodeKey == node3 {
			node3Count += 1

		}
	}
	return node1Count, node2Count, node3Count
}

func TestHashRing(t *testing.T) {
	nodeWeight := make(map[string]int)
	nodeWeight[node1] = 2
	nodeWeight[node2] = 2
	nodeWeight[node3] = 3
	vitualSpots := 100

	hring := New(vitualSpots)
	hring.AddNodes(nodeWeight)
	_, _, c3 := getNodesCount(hring.nodes)

	func() {
		if hring.GetNode("1") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("1"))
		}
		if hring.GetNode("2") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("2"))
		}
		if hring.GetNode("3") != node2 {
			t.Fatalf("expetcd %v got %v", node2, hring.GetNode("3"))
		}
	}()

	func() {
		hring.RemoveNode(node3)
		if hring.GetNode("1") != node1 {
			t.Fatalf("expetcd %v got %v", node1, hring.GetNode("1"))
		}
		if hring.GetNode("2") != node2 {
			t.Fatalf("expetcd %v got %v", node1, hring.GetNode("2"))
		}
		if hring.GetNode("3") != node2 {
			t.Fatalf("expetcd %v got %v", node2, hring.GetNode("3"))
		}
		_, _, _c3 := getNodesCount(hring.nodes)
		assert.Equal(t, 0, _c3)
	}()

	func() {
		hring.AddNode(node3, 3)
		if hring.GetNode("1") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("1"))
		}
		if hring.GetNode("2") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("2"))
		}
		if hring.GetNode("3") != node2 {
			t.Fatalf("expetcd %v got %v", node2, hring.GetNode("3"))
		}
		_, _, _c3 := getNodesCount(hring.nodes)
		assert.Equal(t, c3, _c3)
	}()
}
