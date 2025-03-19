package validate

import (
	"log"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/parser"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var varStrAPI string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "validate a api file",
		Run: func(_ *cobra.Command, _ []string) {
			spec, err := parser.Parse(varStrAPI, "")
			if err != nil {
				log.Fatal(err)
			}

			if err := spec.Validate(); err != nil {
				log.Printf("validate fail: %v\n", err)
			} else {
				log.Print("validate ok")
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&varStrAPI, "api", "", "the api file")
	cmd.MarkFlagRequired("api")
	return cmd
}
