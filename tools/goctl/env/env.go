package env

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/cocktail828/go-tools/tools/goctl/internal/env"
)

func write(_ *cobra.Command, args []string) error {
	if len(sliceVarWriteValue) > 0 {
		return env.WriteEnv(sliceVarWriteValue)
	}
	fmt.Println(env.Print(args...))
	return nil
}
