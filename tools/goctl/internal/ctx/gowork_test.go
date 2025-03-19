package ctx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/rpc/execx"
	"github.com/cocktail828/go-tools/z/stringx"
	"github.com/stretchr/testify/assert"
)

func Test_isGoWork(t *testing.T) {
	dir := filepath.Join("/tmp", stringx.RandomName())

	err := pathx.MkdirIfNotExist(dir)
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	gowork, err := isGoWork(dir)
	assert.False(t, gowork)
	assert.Nil(t, err)

	_, err = execx.Run("go work init", dir)
	assert.Nil(t, err)

	gowork, err = isGoWork(dir)
	assert.True(t, gowork)
	assert.Nil(t, err)

	subDir := filepath.Join(dir, stringx.RandomName())
	err = pathx.MkdirIfNotExist(subDir)
	assert.Nil(t, err)

	gowork, err = isGoWork(subDir)
	assert.True(t, gowork)
	assert.Nil(t, err)
}
