package diagnostic_test

import (
	"io"
	"testing"

	"github.com/cocktail828/go-tools/diagnostic"
	"github.com/stretchr/testify/assert"
)

func TestDiagHasError(t *testing.T) {
	d := diagnostic.New()
	assert.Equal(t, false, d.HasError())
	d = d.WithMessagef("aslkdjfh")
	assert.Equal(t, true, d.HasError())
}

func TestDiagIs(t *testing.T) {
	d := diagnostic.New()
	d = d.WithError(io.ErrClosedPipe)
	assert.Equal(t, true, d.Is(io.ErrClosedPipe))
}

type Err struct{ e string }

func (e Err) Error() string { return e.e }
func TestDiagAs(t *testing.T) {
	err := Err{"lakjsdf"}
	d := diagnostic.New()
	d = d.WithError(err)
	assert.Equal(t, true, d.As(&Err{}))
}

func TestDiagReturnValue(t *testing.T) {
	assert.Equal(t, nil, func() error { return diagnostic.New().ToError() }())
}
