package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/docgen"
	"github.com/cocktail828/go-tools/tools/goctl/format"
	"github.com/cocktail828/go-tools/tools/goctl/gogen"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
	"github.com/cocktail828/go-tools/tools/goctl/validate"
	"github.com/spf13/cobra"
)

const (
	codeFailure = 1
	dash        = "-"
	doubleDash  = "--"
	assign      = "="
)

var (
	builtinFlags = map[string]bool{
		"version": true,
		"help":    true,
	}

	rootCmd = &cobra.Command{
		Use:   "goctl",
		Short: "A versatile tool for Go project generation, API formatting, documentation generation, and more",
	}
)

func init() {
	rootCmd.Version = fmt.Sprintf("%s %s/%s", version.BuildVersion, runtime.GOOS, runtime.GOARCH)
	rootCmd.AddCommand(
		gogen.Command(),
		docgen.Command(),
		format.Command(),
		validate.Command(),
	)
	
	log.SetFlags(log.Lmsgprefix)
	log.SetPrefix("[goctl]: ")
}

// normalizeFlag converts single-dash flags to double-dash format for compatibility
func normalizeFlag(flagExpr string) string {
	flagName, flagValue := flagExpr, ""
	if assignIndex := strings.Index(flagExpr, assign); assignIndex > 0 {
		flagName = flagExpr[:assignIndex]
		flagValue = flagExpr[assignIndex:]
	}

	if !builtinFlags[flagName] {
		return doubleDash + flagName + flagValue
	}
	return dash + flagName + flagValue
}

// supportGoStdFlag ensures compatibility with both single and double dash flags
func supportGoStdFlag(args []string) []string {
	if len(args) == 0 {
		return args
	}

	copyArgs := make([]string, len(args))
	copy(copyArgs, args)

	parentCmd := rootCmd
	for idx, arg := range copyArgs {
		if !strings.HasPrefix(arg, dash) {
			if cmd, _, err := parentCmd.Traverse([]string{arg}); err == nil {
				parentCmd = cmd
			}
			continue
		}

		flagExpr := strings.TrimPrefix(arg, doubleDash)
		flagExpr = strings.TrimPrefix(flagExpr, dash)
		copyArgs[idx] = normalizeFlag(flagExpr)
	}

	return copyArgs
}

func main() {
	rootCmd.SetArgs(supportGoStdFlag(os.Args[1:]))
	if err := rootCmd.Execute(); err != nil {
		log.Printf("Command execution failed: %v", err)
		os.Exit(codeFailure)
	}
}