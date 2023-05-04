package hashring

import (
	"crypto/sha1"
	"sync"

	//	"hash"
	"math"
	"sort"
	"strconv"
)

const (
	//DefaultVirualSpots default virual spots
	DefaultVirualSpots = 400
)

type node struct {
	nodeKey   string
	spotValue uint32
}

type nodesArray []node

func (p nodesArray) Len() int           { return len(p) }
func (p nodesArray) Less(i, j int) bool { return p[i].spotValue < p[j].spotValue }
func (p nodesArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodesArray) Sort()              { sort.Sort(p) }

// HashRing store nodes and weigths
type HashRing struct {
	mu          sync.RWMutex
	virualSpots int
	nodes       nodesArray
	weights     map[string]int
}

// NewHashRing create a hash ring with virual spots
func New(spots int) *HashRing {
	if spots <= 0 {
		spots = DefaultVirualSpots
	}

	return &HashRing{
		virualSpots: spots,
		weights:     make(map[string]int),
	}
}

// AddNode add node to hash ring
func (h *HashRing) AddNode(nodeKey string, weight int) {
	h.AddNodes(map[string]int{nodeKey: weight})
}

// AddNodes add nodes to hash ring
func (h *HashRing) AddNodes(nodeWeight map[string]int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for nodeKey, w := range nodeWeight {
		h.weights[nodeKey] = w
	}
	h.updateLocked()
}

// UpdateNode update node with weight
func (h *HashRing) UpdateNode(nodeKey string, weight int) {
	h.AddNodes(map[string]int{nodeKey: weight})
}

// RemoveNode remove node
func (h *HashRing) RemoveNode(nodeKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.weights, nodeKey)
	h.updateLocked()
}

func (h *HashRing) updateLocked() {
	var totalW int
	for _, w := range h.weights {
		totalW += w
	}

	totalVirtualSpots := h.virualSpots * len(h.weights)
	h.nodes = h.nodes[:0]

	for nodeKey, w := range h.weights {
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(totalVirtualSpots)))
		for i := 1; i <= spots; i++ {
			h.nodes = append(h.nodes, node{
				nodeKey:   nodeKey,
				spotValue: spotValue(nodeKey + ":" + strconv.Itoa(i)),
			})
		}
	}
	h.nodes.Sort()
}

func spotValue(nodeKey string) uint32 {
	hash := sha1.New()
	hash.Write([]byte(nodeKey))
	hashBytes := hash.Sum(nil)[6:10]
	return (uint32(hashBytes[3]) << 24) | (uint32(hashBytes[2]) << 16) | (uint32(hashBytes[1]) << 8) | (uint32(hashBytes[0]))
}

// GetNode get node with key
func (h *HashRing) GetNode(s string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.nodes) == 0 {
		return ""
	}

	v := spotValue(s)
	i := sort.Search(len(h.nodes), func(i int) bool { return h.nodes[i].spotValue >= v })
	if i == len(h.nodes) {
		i = 0
	}
	return h.nodes[i].nodeKey
}
