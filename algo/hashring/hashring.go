package hashring

import (
	"crypto/md5"
	"crypto/sha256"
	"hash/crc32"
	"math"
	"sort"
	"strconv"
	"sync"
)

const (
	//DefaultVirualSpots default virual spots
	DefaultVirualSpots = 300
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
	hashFunc    HashFunc
	virualSpots int
	mu          sync.RWMutex
	nodes       nodesArray
	weights     map[string]int
}

type HashFunc func(string) uint32

func Crc32(nodeKey string) uint32 {
	return crc32.ChecksumIEEE([]byte(nodeKey))
}

func Sha256(nodeKey string) uint32 {
	hash := sha256.New()
	hash.Write([]byte(nodeKey))
	hashBytes := hash.Sum(nil)[6:10]
	return (uint32(hashBytes[3]) << 24) | (uint32(hashBytes[2]) << 16) | (uint32(hashBytes[1]) << 8) | (uint32(hashBytes[0]))
}

func Md5(nodeKey string) uint32 {
	hash := md5.New()
	hash.Write([]byte(nodeKey))
	hashBytes := hash.Sum(nil)[6:10]
	return (uint32(hashBytes[3]) << 24) | (uint32(hashBytes[2]) << 16) | (uint32(hashBytes[1]) << 8) | (uint32(hashBytes[0]))
}

type Option func(*HashRing)

// set hash func
func WithHash(f HashFunc) Option {
	return func(hr *HashRing) {
		hr.hashFunc = f
	}
}

// set num of virtual nodes
func WithSpots(n int) Option {
	return func(hr *HashRing) {
		hr.virualSpots = n
	}
}

// NewHashRing create a hash ring with virual spots
// default Hash crc32
func New(opts ...Option) *HashRing {
	hring := &HashRing{
		hashFunc:    Crc32,
		virualSpots: DefaultVirualSpots,
		weights:     make(map[string]int),
	}

	for _, f := range opts {
		f(hring)
	}

	return hring
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
	h.AddNode(nodeKey, weight)
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
				spotValue: h.hashFunc(nodeKey + ":" + strconv.Itoa(i)),
			})
		}
	}
	h.nodes.Sort()
}

// GetNode get node with key
func (h *HashRing) GetNode(s string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.nodes) == 0 {
		return ""
	}

	v := h.hashFunc(s)
	i := sort.Search(len(h.nodes), func(i int) bool { return h.nodes[i].spotValue >= v })
	if i >= len(h.nodes) || i < 0 {
		i = 0
	}
	return h.nodes[i].nodeKey
}
