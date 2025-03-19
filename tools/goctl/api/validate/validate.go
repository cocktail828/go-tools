package validate

import (
	"errors"

	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/cocktail828/go-tools/tools/goctl/api/parser"
	"github.com/spf13/cobra"
)

// VarStringAPI describes an API.
var VarStringAPI string

// GoValidateApi verifies whether the api has a syntax error
func GoValidateApi(_ *cobra.Command, _ []string) error {
	apiFile := VarStringAPI

	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}

	spec, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	err = spec.Validate()
	if err == nil {
		colorful.Infoln("api format ok")
	}
	return err
}
