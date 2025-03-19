package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	sortedmap "github.com/cocktail828/go-tools/tools/goctl/internal/collection"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
	"github.com/cocktail828/go-tools/tools/goctl/vars"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/pkg/errors"
)

var goctlEnv *sortedmap.SortedMap

const (
	GoctlOS                = "GOCTL_OS"
	GoctlArch              = "GOCTL_ARCH"
	GoctlHome              = "GOCTL_HOME"
	GoctlDebug             = "GOCTL_DEBUG"
	GoctlCache             = "GOCTL_CACHE"
	GoctlVersion           = "GOCTL_VERSION"
	ProtocVersion          = "PROTOC_VERSION"
	ProtocGenGoVersion     = "PROTOC_GEN_GO_VERSION"
	ProtocGenGoGRPCVersion = "PROTO_GEN_GO_GRPC_VERSION"

	envFileDir      = "env"
	ExperimentalOn  = "on"
	ExperimentalOff = "off"
)

// init initializes the goctl environment variables, the environment variables of the function are set in order,
// please do not change the logic order of the code.
func init() {
	defaultGoctlHome, err := pathx.GetDefaultGoctlHome()
	if err != nil {
		colorful.Fatalln(err)
	}
	goctlEnv = sortedmap.New()
	goctlEnv.SetKV(GoctlOS, runtime.GOOS)
	goctlEnv.SetKV(GoctlArch, runtime.GOARCH)
	existsEnv := readEnv(defaultGoctlHome)
	if existsEnv != nil {
		goctlHome, ok := existsEnv.GetString(GoctlHome)
		if ok && len(goctlHome) > 0 {
			goctlEnv.SetKV(GoctlHome, goctlHome)
		}
		if debug := existsEnv.GetOr(GoctlDebug, "").(string); debug != "" {
			if strings.EqualFold(debug, "true") || strings.EqualFold(debug, "false") {
				goctlEnv.SetKV(GoctlDebug, debug)
			}
		}
		if value := existsEnv.GetStringOr(GoctlCache, ""); value != "" {
			goctlEnv.SetKV(GoctlCache, value)
		}
	}

	if !goctlEnv.HasKey(GoctlHome) {
		goctlEnv.SetKV(GoctlHome, defaultGoctlHome)
	}
	if !goctlEnv.HasKey(GoctlDebug) {
		goctlEnv.SetKV(GoctlDebug, "False")
	}

	if !goctlEnv.HasKey(GoctlCache) {
		cacheDir, _ := pathx.GetCacheDir()
		goctlEnv.SetKV(GoctlCache, cacheDir)
	}

	goctlEnv.SetKV(GoctlVersion, version.BuildVersion)

	kvs := map[string]Command{
		ProtocVersion:          Protoc{},
		ProtocGenGoVersion:     ProtocGenGo{},
		ProtocGenGoGRPCVersion: ProtocGenGoGrpc{},
	}

	for k, v := range kvs {
		ver, _ := v.Version()
		goctlEnv.SetKV(k, ver)
	}
}

func Print(args ...string) string {
	if len(args) == 0 {
		return strings.Join(goctlEnv.Format(), "\n")
	}

	var values []string
	for _, key := range args {
		value, ok := goctlEnv.GetString(key)
		if !ok {
			value = fmt.Sprintf("%s=%%not found%%", key)
		}
		values = append(values, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(values, "\n")
}

func Get(key string) string {
	return GetOr(key, "")
}

// Set sets the environment variable for testing
func Set(t *testing.T, key, value string) {
	goctlEnv.SetKV(key, value)
	t.Cleanup(func() {
		goctlEnv.Remove(key)
	})
}

func GetOr(key, def string) string {
	return goctlEnv.GetStringOr(key, def)
}

func readEnv(goctlHome string) *sortedmap.SortedMap {
	envFile := filepath.Join(goctlHome, envFileDir)
	data, err := os.ReadFile(envFile)
	if err != nil {
		return nil
	}
	dataStr := string(data)
	lines := strings.Split(dataStr, "\n")
	sm := sortedmap.New()
	for _, line := range lines {
		_, _, err = sm.SetExpression(line)
		if err != nil {
			continue
		}
	}
	return sm
}

func WriteEnv(kv []string) error {
	defaultGoctlHome, err := pathx.GetDefaultGoctlHome()
	if err != nil {
		colorful.Fatalln(err)
	}
	data := sortedmap.New()
	for _, e := range kv {
		_, _, err := data.SetExpression(e)
		if err != nil {
			return err
		}
	}
	data.RangeIf(func(key, value any) bool {
		switch key.(string) {
		case GoctlHome, GoctlCache:
			path := value.(string)
			if !pathx.FileExists(path) {
				err = errors.Errorf("[writeEnv]: path %q is not exists", path)
				return false
			}
		}
		if goctlEnv.HasKey(key) {
			goctlEnv.SetKV(key, value)
			return true
		} else {
			err = errors.Errorf("[writeEnv]: invalid key: %v", key)
			return false
		}
	})
	if err != nil {
		return err
	}
	envFile := filepath.Join(defaultGoctlHome, envFileDir)
	return os.WriteFile(envFile, []byte(strings.Join(goctlEnv.Format(), "\n")), 0o777)
}

const (
	bin                = "bin"
	binGo              = "go"
	binProtoc          = "protoc"
	binProtocGenGo     = "protoc-gen-go"
	binProtocGenGrpcGo = "protoc-gen-go-grpc"
	cstOffset          = 60 * 60 * 8 // 8 hours offset for Chinese Standard Time
)

// InChina returns whether the current time is in China Standard Time.
func InChina() bool {
	_, offset := time.Now().Zone()
	return offset == cstOffset
}

// LookUpGo searches an executable go in the directories
// named by the GOROOT/bin or PATH environment variable.
func LookUpGo() (string, error) {
	goRoot := runtime.GOROOT()
	suffix := getExeSuffix()
	xGo := binGo + suffix
	path := filepath.Join(goRoot, bin, xGo)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return LookPath(xGo)
}

// LookUpProtoc searches an executable protoc in the directories
// named by the PATH environment variable.
func LookUpProtoc() (string, error) {
	suffix := getExeSuffix()
	xProtoc := binProtoc + suffix
	return LookPath(xProtoc)
}

// LookUpProtocGenGo searches an executable protoc-gen-go in the directories
// named by the PATH environment variable.
func LookUpProtocGenGo() (string, error) {
	suffix := getExeSuffix()
	xProtocGenGo := binProtocGenGo + suffix
	return LookPath(xProtocGenGo)
}

// LookUpProtocGenGoGrpc searches an executable protoc-gen-go-grpc in the directories
// named by the PATH environment variable.
func LookUpProtocGenGoGrpc() (string, error) {
	suffix := getExeSuffix()
	xProtocGenGoGrpc := binProtocGenGrpcGo + suffix
	return LookPath(xProtocGenGoGrpc)
}

// LookPath searches for an executable named file in the
// directories named by the PATH environment variable,
// for the os windows, the named file will be spliced with the
// .exe suffix.
func LookPath(xBin string) (string, error) {
	suffix := getExeSuffix()
	if len(suffix) > 0 && !strings.HasSuffix(xBin, suffix) {
		xBin = xBin + suffix
	}

	bin, err := exec.LookPath(xBin)
	if err != nil {
		return "", err
	}
	return bin, nil
}

func getExeSuffix() string {
	if runtime.GOOS == vars.OsWindows {
		return ".exe"
	}
	return ""
}
