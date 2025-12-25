package z

import (
	"errors"
	"fmt"
	"strconv"
)

// Memory represent memory size, the underlying unit is byte.
type Memory int64

// Constants define memory size based on 1024 binary units (IEC standard).
const (
	Byte     Memory = 1
	KiloByte        = Byte * 1024
	MegaByte        = KiloByte * 1024
	GigaByte        = MegaByte * 1024
	TeraByte        = GigaByte * 1024
	PetaByte        = TeraByte * 1024
)

// Bytes return the size in bytes.
func (s Memory) Bytes() int64       { return int64(s) }
func (s Memory) Kilobytes() float64 { return float64(s) / float64(KiloByte) }
func (s Memory) Megabytes() float64 { return float64(s) / float64(MegaByte) }
func (s Memory) Gigabytes() float64 { return float64(s) / float64(GigaByte) }
func (s Memory) Terabytes() float64 { return float64(s) / float64(TeraByte) }

var memoryUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

// String return the size in a human-readable format (e.g. "1.5 GB").
func (s Memory) String() string {
	if s == 0 {
		return "0B"
	}

	bytes := float64(s)
	unitIndex := 0

	for bytes >= 1024 && unitIndex < len(memoryUnits)-1 {
		bytes /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d%s", int64(s), memoryUnits[unitIndex])
	}

	return fmt.Sprintf("%.2f%s", bytes, memoryUnits[unitIndex])
}

var errLeadingInt = errors.New("memory: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt[bytes []byte | string](s bytes) (x uint64, rem bytes, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, rem, errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, rem, errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

var unitMap = map[string]uint64{
	"B":  uint64(Byte),
	"KB": uint64(KiloByte),
	"MB": uint64(MegaByte),
	"GB": uint64(GigaByte),
	"TB": uint64(TeraByte),
	"PB": uint64(PetaByte),
	"K":  uint64(KiloByte),
	"M":  uint64(MegaByte),
	"G":  uint64(GigaByte),
	"T":  uint64(TeraByte),
	"P":  uint64(PetaByte),
}

func quote(s string) string { return strconv.Quote(s) }

// ParseMemory parses a string representing a memory size into bytes.
// Supported units (case-insensitive): B, KB, MB, GB, TB, PB (binary, base-1024).
// Examples:
//
//	"10B"    → 10
//	"512KB"  → 512 * 1024
//	"2.5GB"  → 2.5 * 1024^3
//	"1TB"    → 1 * 1024^4
func ParseMemory(s string) (Memory, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	if neg {
		return 0, errors.New("memory cannot be negative")
	}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("memory: invalid memory " + quote(orig))
	}
	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, errors.New("memory: invalid memory " + quote(orig))
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New("memory: invalid memory " + quote(orig))
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New("memory: invalid memory " + quote(orig))
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("memory: missing unit in memory " + quote(orig))
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("memory: unknown unit " + quote(u) + " in memory " + quote(orig))
		}
		if v > 1<<63/unit {
			// overflow
			return 0, errors.New("memory: invalid memory " + quote(orig))
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, errors.New("memory: invalid memory " + quote(orig))
			}
		}
		d += v
		if d > 1<<63 {
			return 0, errors.New("memory: invalid memory " + quote(orig))
		}
	}
	if d > 1<<63-1 {
		return 0, errors.New("memory: invalid memory " + quote(orig))
	}
	return Memory(d), nil
}
