package cm4

import (
	"testing"

	"github.com/cocktail828/go-tools/z/mathx"
	"github.com/stretchr/testify/require"
)

func TestCM4(t *testing.T) {
	defer func() {
		require.NotNil(t, recover())
	}()

	s := NewCM4(5)
	require.Equal(t, uint64(7), s.mask)
	NewCM4(0)
}

func TestCM4Increment(t *testing.T) {
	s := NewCM4(16)
	s.Increment(1)
	s.Increment(5)
	s.Increment(9)
	for i := 0; i < cmDepth; i++ {
		if s.rows[i].String() != s.rows[0].String() {
			break
		}
		require.False(t, i == cmDepth-1, "identical rows, bad seeding")
	}
}

func TestCM4Estimate(t *testing.T) {
	s := NewCM4(16)
	s.Increment(1)
	s.Increment(1)
	require.Equal(t, int64(2), s.Estimate(1))
	require.Equal(t, int64(0), s.Estimate(0))
}

func TestCM4Reset(t *testing.T) {
	s := NewCM4(16)
	s.Increment(1)
	s.Increment(1)
	s.Increment(1)
	s.Increment(1)
	s.Reset()
	require.Equal(t, int64(2), s.Estimate(1))
}

func TestCM4Clear(t *testing.T) {
	s := NewCM4(16)
	for i := 0; i < 16; i++ {
		s.Increment(uint64(i))
	}
	s.Clear()
	for i := 0; i < 16; i++ {
		require.Equal(t, int64(0), s.Estimate(uint64(i)))
	}
}

func TestNext2Power(t *testing.T) {
	sz := 12 << 30
	szf := float64(sz) * 0.01
	val := int64(szf)
	t.Logf("szf = %.2f val = %d\n", szf, val)
	pow := mathx.Next2Power(val)
	t.Logf("pow = %d. mult 4 = %d\n", pow, pow*4)
}

func BenchmarkCM4Increment(b *testing.B) {
	s := NewCM4(16)
	b.SetBytes(1)
	for n := 0; n < b.N; n++ {
		s.Increment(1)
	}
}

func BenchmarkCM4Estimate(b *testing.B) {
	s := NewCM4(16)
	s.Increment(1)
	b.SetBytes(1)
	for n := 0; n < b.N; n++ {
		s.Estimate(1)
	}
}
