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
	assert.Equal(t, "testaaa", AsString(envname, StringVal("testaaa")))
	t.Setenv(envname, "test")
	assert.Equal(t, "test", AsString(envname))
}

func TestInt(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.EqualValues(t, 10, AsInt(envname, IntVal(10)))
	t.Setenv(envname, "1")
	assert.EqualValues(t, 1, AsInt(envname))
}

func TestBool(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.Equal(t, true, AsBool(envname, BoolVal(true)))
	t.Setenv(envname, "1")
	assert.Equal(t, true, AsBool(envname))
}

func TestFloat64(t *testing.T) {
	envname := "XTEST"
	defer t.Cleanup(func() {
		os.Setenv(envname, "")
	})
	assert.Equal(t, 10.0, AsFloat(envname, FloatVal(10.0)))
	t.Setenv(envname, "1")
	assert.Equal(t, 1.0, AsFloat(envname))
}
