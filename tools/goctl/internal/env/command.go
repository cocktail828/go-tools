package env

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cocktail828/go-tools/tools/goctl/internal/golang"
	"github.com/cocktail828/go-tools/tools/goctl/rpc/execx"
	"github.com/cocktail828/go-tools/tools/goctl/vars"
	"github.com/pkg/errors"
)

type Command interface {
	Name() string
	Install(cacheDir string) (string, error)
	Exists() bool
	Version() (string, error)
}

type Protoc struct{}

func (c Protoc) Name() string { return "protoc" }
func (c Protoc) Install(cacheDir string) (string, error) {
	zipFileName := c.Name() + ".zip"
	var url = map[string]string{
		"linux_32":   "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_32.zip",
		"linux_64":   "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_64.zip",
		"darwin":     "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-osx-x86_64.zip",
		"windows_32": "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-win32.zip",
		"windows_64": "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-win64.zip",
	}

	return Install(cacheDir, c.Name(), func(dest string) (string, error) {
		goos := runtime.GOOS
		tempFile := filepath.Join(os.TempDir(), zipFileName)
		bit := 32 << (^uint(0) >> 63)
		var downloadUrl string
		switch goos {
		case vars.OsMac:
			downloadUrl = url[vars.OsMac]
		case vars.OsWindows:
			downloadUrl = url[fmt.Sprintf("%s_%d", vars.OsWindows, bit)]
		case vars.OsLinux:
			downloadUrl = url[fmt.Sprintf("%s_%d", vars.OsLinux, bit)]
		default:
			return "", errors.Errorf("unsupport OS: %q", goos)
		}

		if err := Download(downloadUrl, tempFile); err != nil {
			return "", err
		}

		return dest, Unpacking(tempFile, filepath.Dir(dest), func(f *zip.File) bool {
			return filepath.Base(f.Name) == filepath.Base(dest)
		})
	})
}

func (c Protoc) Exists() bool {
	_, err := LookUpProtoc()
	return err == nil
}

func (c Protoc) Version() (string, error) {
	path, err := LookUpProtoc()
	if err != nil {
		return "", err
	}
	version, err := execx.Run(path+" --version", "")
	if err != nil {
		return "", err
	}
	fields := strings.Fields(version)
	if len(fields) > 1 {
		return fields[1], nil
	}
	return "", nil
}

type ProtocGenGo struct{}

func (c ProtocGenGo) Name() string { return "protoc-gen-go" }
func (c ProtocGenGo) Install(cacheDir string) (string, error) {
	url := "google.golang.org/protobuf/cmd/protoc-gen-go@latest"

	return Install(cacheDir, c.Name(), func(dest string) (string, error) {
		err := golang.Install(url)
		return dest, err
	})
}

func (c ProtocGenGo) Exists() bool {
	ver, err := c.Version()
	if err != nil {
		return false
	}
	return len(ver) > 0
}

// Version is used to get the version of the protoc-gen-go plugin. For older versions, protoc-gen-go does not support
// version fetching, so if protoc-gen-go --version is executed, it will cause the process to block, so it is controlled
// by a timer to prevent the older version process from blocking.
func (c ProtocGenGo) Version() (string, error) {
	path, err := LookUpProtocGenGo()
	if err != nil {
		return "", err
	}
	versionC := make(chan string)
	go func(c chan string) {
		version, _ := execx.Run(path+" --version", "")
		fields := strings.Fields(version)
		if len(fields) > 1 {
			c <- fields[1]
		}
	}(versionC)
	t := time.NewTimer(time.Second)
	select {
	case <-t.C:
		return "", nil
	case version := <-versionC:
		return version, nil
	}
}

type ProtocGenGoGrpc struct{}

func (c ProtocGenGoGrpc) Name() string { return "protoc-gen-go-grpc" }
func (c ProtocGenGoGrpc) Install(cacheDir string) (string, error) {
	url := "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"

	return Install(cacheDir, c.Name(), func(dest string) (string, error) {
		err := golang.Install(url)
		return dest, err
	})
}

func (c ProtocGenGoGrpc) Exists() bool {
	_, err := LookUpProtocGenGoGrpc()
	return err == nil
}

// Version is used to get the version of the protoc-gen-go-grpc plugin.
func (c ProtocGenGoGrpc) Version() (string, error) {
	path, err := LookUpProtocGenGoGrpc()
	if err != nil {
		return "", err
	}
	version, err := execx.Run(path+" --version", "")
	if err != nil {
		return "", err
	}
	fields := strings.Fields(version)
	if len(fields) > 1 {
		return fields[1], nil
	}
	return "", nil
}
