package gogen

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed main.tpl
var mainTpl []byte

func TestLoadTemplate(t *testing.T) {
	tpl := Template{}

	tests := []struct {
		relativepath string
		fname        string
		expt         []byte
	}{
		{"", "main.tpl", mainTpl},
		{".", "main.tpl", mainTpl},
		{"./", "main.tpl", mainTpl},
		{".//", "main.tpl", mainTpl},
		{"./xxx", "main.tpl", mainTpl},
	}

	for _, tt := range tests {
		payload, err := tpl.Load(tt.relativepath, tt.fname)
		require.NoError(t, err)
		require.EqualValuesf(t, tt.expt, payload, "relativepath:%s, fname:%s", tt.relativepath, tt.fname)
	}
}
