package environ

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.Equal(t, "testaaa", String(envname, WithString("testaaa")))
	t.Setenv(envname, "test")
	assert.Equal(t, "test", String(envname))
}

func TestInt(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.EqualValues(t, 10, Int(envname, WithInt(10)))
	t.Setenv(envname, "1")
	assert.EqualValues(t, 1, Int(envname))
}

func TestBool(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.Equal(t, true, Bool(envname, WithBool(true)))
	t.Setenv(envname, "1")
	assert.Equal(t, true, Bool(envname))
}

func TestFloat64(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.Equal(t, 10.0, Float(envname, WithFloat(10.0)))
	t.Setenv(envname, "1")
	assert.Equal(t, 1.0, Float(envname))
}
