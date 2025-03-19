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
	"github.com/cocktail828/go-tools/tools/goctl/api/parser"
	apiutil "github.com/cocktail828/go-tools/tools/goctl/api/util"
	"github.com/cocktail828/go-tools/tools/goctl/internal/golang"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/cocktail828/go-tools/z"
	"github.com/spf13/cobra"
)

const tmpFile = "%s-%d"

const (
	typesPacket   = "model"
	configDir     = "config"
	serviceDir    = "service"
	handlerDir    = "handler"
	logicDir      = "logic"
	middlewareDir = "middleware"
	typesDir      = typesPacket
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
	// VarStringRemote describes the remote git repository.
	VarStringRemote string
	// VarStringBranch describes the branch.
	VarStringBranch string
)

// GoCommand gen go project files from command line
func GoCommand(_ *cobra.Command, _ []string) error {
	apiFile := VarStringAPI
	dir := VarStringDir
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}

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
	api, err := parser.Parse(apiFile)
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
	z.Must(genLogic(dir, rootPkg, api))
	z.Must(genMiddleware(dir, api))
	if withTest {
		z.Must(genHandlersTest(dir, rootPkg, api))
		z.Must(genLogicTest(dir, rootPkg, api))
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
