package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/api"
	"github.com/cocktail828/go-tools/tools/goctl/env"
	"github.com/cocktail828/go-tools/tools/goctl/internal/cobrax"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
	"github.com/cocktail828/go-tools/tools/goctl/rpc"
	"github.com/cocktail828/go-tools/tools/goctl/tpl"
	"github.com/cocktail828/go-tools/z"
)

const (
	codeFailure = 1
	dash        = "-"
	doubleDash  = "--"
	assign      = "="
)

var (
	rootCmd = cobrax.NewCommand("goctl")
)

func init() {
	rootCmd.Version = fmt.Sprintf("%s %s/%s", version.BuildVersion, runtime.GOOS, runtime.GOARCH)
	rootCmd.AddCommand(api.Cmd, env.Cmd, rpc.Cmd, tpl.Cmd)
	rootCmd.MustInit()

	log.SetFlags(0)
}

func isBuiltin(name string) bool {
	return name == "version" || name == "help"
}

func supportGoStdFlag(args []string) []string {
	copyArgs := append([]string(nil), args...)
	parentCmd, _, err := rootCmd.Traverse(args[:1])
	if err != nil { // ignore it to let cobra handle the error.
		return copyArgs
	}

	for idx, arg := range copyArgs[0:] {
		parentCmd, _, err = parentCmd.Traverse([]string{arg})
		if err != nil { // ignore it to let cobra handle the error.
			break
		}
		if !strings.HasPrefix(arg, dash) {
			continue
		}

		flagExpr := strings.TrimPrefix(arg, doubleDash)
		flagExpr = strings.TrimPrefix(flagExpr, dash)
		flagName, flagValue := flagExpr, ""
		assignIndex := strings.Index(flagExpr, assign)
		if assignIndex > 0 {
			flagName = flagExpr[:assignIndex]
			flagValue = flagExpr[assignIndex:]
		}

		if !isBuiltin(flagName) {
			// The method Flag can only match the user custom flags.
			f := parentCmd.Flag(flagName)
			if f == nil {
				continue
			}
			if f.Shorthand == flagName {
				continue
			}
		}

		goStyleFlag := doubleDash + flagName
		if assignIndex > 0 {
			goStyleFlag += flagValue
		}

		copyArgs[idx] = goStyleFlag
	}
	return copyArgs
}

func main() {
	os.Args = supportGoStdFlag(os.Args)
	z.Must(rootCmd.Execute())
}
