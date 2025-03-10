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
	key       string
	spotValue uint32
}

type nodeArray []node

func (p nodeArray) Len() int           { return len(p) }
func (p nodeArray) Less(i, j int) bool { return p[i].spotValue < p[j].spotValue }
func (p nodeArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodeArray) Sort()              { sort.Sort(p) }

// HashRing store nodes and weigths
type HashRing struct {
	hashFunc    HashFunc
	virualSpots int
	mu          sync.RWMutex
	nodes       nodeArray
	weights     map[string]int
}

type HashFunc func(string) uint32

func Crc32(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func Sha256(key string) uint32 {
	hash := sha256.New()
	hash.Write([]byte(key))
	hashBytes := hash.Sum(nil)[6:10]
	return (uint32(hashBytes[3]) << 24) | (uint32(hashBytes[2]) << 16) | (uint32(hashBytes[1]) << 8) | (uint32(hashBytes[0]))
}

func Md5(key string) uint32 {
	hash := md5.New()
	hash.Write([]byte(key))
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
func WithVirtualSpots(n int) Option {
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

func (h *HashRing) AddMany(nodeWeight map[string]int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, w := range nodeWeight {
		h.weights[key] = w
	}
	h.updateLocked()
}

func (h *HashRing) Add(key string, weight int) {
	h.AddMany(map[string]int{key: weight})
}

func (h *HashRing) Update(key string, weight int) {
	h.Add(key, weight)
}

func (h *HashRing) Remove(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.weights, key)
	h.updateLocked()
}

func (h *HashRing) updateLocked() {
	var totalW int
	for _, w := range h.weights {
		totalW += w
	}

	totalVirtualSpots := h.virualSpots * len(h.weights)
	h.nodes = h.nodes[:0]

	for key, w := range h.weights {
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(totalVirtualSpots)))
		for i := 1; i <= spots; i++ {
			h.nodes = append(h.nodes, node{
				key:       key,
				spotValue: h.hashFunc(key + ":" + strconv.Itoa(i)),
			})
		}
	}
	h.nodes.Sort()
}

func (h *HashRing) Get(s string) string {
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
	return h.nodes[i].key
}
