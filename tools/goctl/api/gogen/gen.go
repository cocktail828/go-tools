package gogen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	apiformat "github.com/cocktail828/go-tools/tools/goctl/api/format"
	apiutil "github.com/cocktail828/go-tools/tools/goctl/api/util"
	"github.com/cocktail828/go-tools/tools/goctl/internal/golang"
	apiParser "github.com/cocktail828/go-tools/tools/goctl/internal/parser/parser"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/cocktail828/go-tools/z"
	"github.com/spf13/cobra"
)

const tmpFile = "%s-%d"

const (
	serviceDir    = "service"
	handlerDir    = "handler"
	middlewareDir = "middleware"
	typesPacket   = "model"
	typesDir      = handlerDir + "/" + typesPacket
	groupProperty = "group"
)

var (
	tmpDir = path.Join(os.TempDir(), "goctl")
	// VarStringDir describes the directory.
	VarStringDir string
	// VarStringAPI describes the API.
	VarStringAPI string
	// VarStringHome describes the go home.
	VarStringHome string
)

// GoCommand gen go project files from command line
func GoCommand(_ *cobra.Command, _ []string) error {
	apiFile := VarStringAPI
	dir := VarStringDir
	home := VarStringHome

	if len(home) > 0 {
		pathx.RegisterGoctlHome(home)
	}
	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	return DoGenProject(apiFile, dir, true)
}

// DoGenProject gen go project files with api file
func DoGenProject(apiFile, dir string, withTest bool) error {
	api, err := apiParser.Parse(apiFile, "")
	if err != nil {
		return err
	}

	if err := api.Validate(); err != nil {
		return err
	}

	z.Must(pathx.MkdirIfNotExist(dir))
	rootPkg, err := golang.GetParentPackage(dir)
	if err != nil {
		return err
	}

	z.Must(genMain(dir, rootPkg, api))
	z.Must(genService(dir, rootPkg, api))
	z.Must(genModel(dir, api))
	z.Must(genRoutes(dir, rootPkg, api))
	z.Must(genHandlers(dir, rootPkg, api))
	z.Must(genMiddleware(dir, api))
	if withTest {
		z.Must(genHandlersTest(dir, rootPkg, api))
	}

	if err := backupAndSweep(apiFile); err != nil {
		return err
	}

	if err := apiformat.ApiFormatByPath(apiFile, false); err != nil {
		return err
	}

	return nil
}

func backupAndSweep(apiFile string) error {
	var err error
	var wg sync.WaitGroup

	wg.Add(2)
	_ = os.MkdirAll(tmpDir, os.ModePerm)

	go func() {
		_, fileName := filepath.Split(apiFile)
		_, e := apiutil.Copy(apiFile, fmt.Sprintf(path.Join(tmpDir, tmpFile), fileName, time.Now().Unix()))
		if e != nil {
			err = e
		}
		wg.Done()
	}()
	go func() {
		if e := sweep(); e != nil {
			err = e
		}
		wg.Done()
	}()
	wg.Wait()

	return err
}

func sweep() error {
	keepTime := time.Now().AddDate(0, 0, -7)
	return filepath.Walk(tmpDir, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		pos := strings.LastIndexByte(info.Name(), '-')
		if pos > 0 {
			timestamp := info.Name()[pos+1:]
			seconds, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				// print error and ignore
				colorful.Warnf("sweep ignored file: %s", fpath)
				return nil
			}

			tm := time.Unix(seconds, 0)
			if tm.Before(keepTime) {
				if err := os.RemoveAll(fpath); err != nil {
					colorful.Warnf("failed to remove file: %s", fpath)
					return err
				}
			}
		}

		return nil
	})
}
