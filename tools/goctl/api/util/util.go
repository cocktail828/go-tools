package util

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/api/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/z"
)

// ShouldCreateFile creates file if not exists
func ShouldCreateFile(dir, subdir, file string) (fp *os.File, created bool, err error) {
	z.Must(pathx.MkdirIfNotExist(path.Join(dir, subdir)))
	fpath := path.Join(dir, subdir, file)
	if pathx.FileExists(fpath) {
		fmt.Printf("%s exists, ignored generation\n", fpath)
		return nil, false, nil
	}

	fp, err = pathx.CreateIfNotExist(fpath)
	created = err == nil
	return
}

// Copy calls io.Copy if the source file and destination file exists
func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// ComponentName returns component name for typescript
func ComponentName(api *spec.ApiSpec) string {
	name := api.Service.Name
	if strings.HasSuffix(name, "-api") {
		return name[:len(name)-4] + "Components"
	}
	return name + "Components"
}

// WriteIndent writes tab spaces
func WriteIndent(writer io.Writer, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(writer, "\t")
	}
}

// RemoveComment filters comment content
func RemoveComment(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		return strings.TrimSpace(line[:commentIdx])
	}
	return strings.TrimSpace(line)
}
