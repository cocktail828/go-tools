// Package snowflake provides a very simple Twitter snowflake generator and parser.
package snowflake

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	// Epoch is set to the Twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds.
	// You may customize this to set a different epoch for your application.
	Epoch int64 = 1288834974657

	// NodeBits holds the number of bits to use for Node.
	// Remember, you have a total of 22 bits to share between Node/Step.
	NodeBits uint8 = 10

	// StepBits holds the number of bits to use for Step.
	// Remember, you have a total of 22 bits to share between Node/Step.
	StepBits uint8 = 12

	// Pre-calculate constants
	maxNode   = -1 ^ (-1 << NodeBits)
	maxStep   = -1 ^ (-1 << StepBits)
	timeShift = NodeBits + StepBits
	nodeShift = StepBits
)

const encodeBase32Map = "ybndrfg8ejkmcpqxot1uwisza345h769"

var decodeBase32Map [256]byte

const encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var decodeBase58Map [256]byte

// ErrInvalidBase58 is returned by ParseBase58 when given an invalid []byte.
var ErrInvalidBase58 = errors.New("invalid base58")

// ErrInvalidBase32 is returned by ParseBase32 when given an invalid []byte.
var ErrInvalidBase32 = errors.New("invalid base32")

// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.
func init() {
	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[encodeBase32Map[i]] = byte(i)
	}
}

// A Node struct holds the basic information needed for a snowflake generator node.
type Node struct {
	mu    sync.Mutex
	epoch time.Time
	time  int64
	node  int64
	step  int64
}

// An ID is a custom type used for a snowflake ID. This is used so we can
// attach methods onto the ID.
type ID int64

// NewNode returns a new snowflake node that can be used to generate snowflake IDs.
func NewNode(node int64) (*Node, error) {
	if node < 0 || node > maxNode {
		return nil, fmt.Errorf("node number must be between 0 and %d", maxNode)
	}

	return &Node{
		epoch: time.UnixMilli(Epoch),
		node:  node,
	}, nil
}

// Generate creates and returns a unique snowflake ID.
// To help guarantee uniqueness:
// - Make sure your system is keeping accurate system time.
// - Make sure you never have multiple nodes running with the same node ID.
func (n *Node) Generate() ID {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Since(n.epoch).Milliseconds()

	if now == n.time {
		n.step = (n.step + 1) & maxStep
		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Milliseconds()
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	return ID((now << timeShift) | (n.node << nodeShift) | (n.step))
}

// Int64 returns an int64 of the snowflake ID.
func (f ID) Int64() int64 {
	return int64(f)
}

// ParseInt64 converts an int64 into a snowflake ID.
func ParseInt64(id int64) ID {
	return ID(id)
}

// String returns a string of the snowflake ID.
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// ParseString converts a string into a snowflake ID.
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse string %q as snowflake ID: %w", id, err)
	}
	return ID(i), nil
}

// Base2 returns a string base2 of the snowflake ID.
func (f ID) Base2() string {
	return strconv.FormatInt(int64(f), 2)
}

// ParseBase2 converts a Base2 string into a snowflake ID.
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base2 string %q as snowflake ID: %w", id, err)
	}
	return ID(i), nil
}

// encode is a helper function for encoding to a custom base.
func (f ID) encode(baseMap string, base int) string {
	if f < ID(base) {
		return string(baseMap[f])
	}

	b := make([]byte, 0, 12)
	for f >= ID(base) {
		b = append(b, baseMap[f%ID(base)])
		f /= ID(base)
	}
	b = append(b, baseMap[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// Base32 uses the z-base-32 character set but encodes and decodes similar
// to base58, allowing it to create an even smaller result string.
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func (f ID) Base32() string {
	return f.encode(encodeBase32Map, 32)
}

// ParseBase32 parses a base32 []byte into a snowflake ID
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func ParseBase32(b []byte) (ID, error) {
	return parseBase(b, decodeBase32Map, 32, ErrInvalidBase32)
}

// Base36 returns a base36 string of the snowflake ID.
func (f ID) Base36() string {
	return strconv.FormatInt(int64(f), 36)
}

// ParseBase36 converts a Base36 string into a snowflake ID.
func ParseBase36(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 36, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base36 string %q as snowflake ID: %w", id, err)
	}
	return ID(i), nil
}

// Base58 returns a base58 string of the snowflake ID.
func (f ID) Base58() string {
	return f.encode(encodeBase58Map, 58)
}

// ParseBase58 parses a base58 []byte into a snowflake ID.
func ParseBase58(b []byte) (ID, error) {
	return parseBase(b, decodeBase58Map, 58, ErrInvalidBase58)
}

// Base64 returns a base64 string of the snowflake ID.
func (f ID) Base64() string {
	return base64.StdEncoding.EncodeToString(f.Bytes())
}

// ParseBase64 converts a base64 string into a snowflake ID.
func ParseBase64(id string) (ID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return -1, fmt.Errorf("failed to decode base64 string %q: %w", id, err)
	}
	return ParseBytes(b)
}

// Bytes returns a byte slice of the snowflake ID.
func (f ID) Bytes() []byte {
	return []byte(f.String())
}

// ParseBytes converts a byte slice into a snowflake ID.
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse byte slice %q as snowflake ID: %w", string(id), err)
	}
	return ID(i), nil
}

// IntBytes returns an array of bytes of the snowflake ID, encoded as a
// big endian integer.
func (f ID) IntBytes() [8]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(f))
	return b
}

// ParseIntBytes converts an array of bytes encoded as big endian integer as
// a snowflake ID.
func ParseIntBytes(id [8]byte) ID {
	return ID(int64(binary.BigEndian.Uint64(id[:])))
}

// MarshalJSON returns a JSON byte array string of the snowflake ID.
func (f ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

// UnmarshalJSON converts a JSON byte array of a snowflake ID into an ID type.
func (f *ID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	id, err := ParseString(s)
	if err != nil {
		return err
	}
	*f = id
	return nil
}

// parseBase is a helper function for parsing from a custom base.
func parseBase(b []byte, decodeMap [256]byte, base int, err error) (ID, error) {
	var id int64
	for i := range b {
		if decodeMap[b[i]] == 0xFF {
			return -1, fmt.Errorf("invalid base%d character %c in input: %w", base, b[i], err)
		}
		id = id*int64(base) + int64(decodeMap[b[i]])
	}
	return ID(id), nil
}
