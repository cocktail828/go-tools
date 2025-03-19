package api

import (
	"github.com/cocktail828/go-tools/tools/goctl/api/docgen"
	"github.com/cocktail828/go-tools/tools/goctl/api/format"
	"github.com/cocktail828/go-tools/tools/goctl/api/gogen"
	"github.com/cocktail828/go-tools/tools/goctl/api/validate"
	"github.com/cocktail828/go-tools/tools/goctl/internal/cobrax"
	"github.com/cocktail828/go-tools/tools/goctl/plugin"
)

var (
	// Cmd describes an api command.
	Cmd         = cobrax.NewCommand("api")
	docCmd      = cobrax.NewCommand("doc", cobrax.WithRunE(docgen.DocCommand))
	formatCmd   = cobrax.NewCommand("format", cobrax.WithRunE(format.GoFormatApi))
	goCmd       = cobrax.NewCommand("go", cobrax.WithRunE(gogen.GoCommand))
	validateCmd = cobrax.NewCommand("validate", cobrax.WithRunE(validate.GoValidateApi))
	pluginCmd   = cobrax.NewCommand("plugin", cobrax.WithRunE(plugin.PluginCommand))
)

func init() {
	var (
		docCmdFlags      = docCmd.Flags()
		formatCmdFlags   = formatCmd.Flags()
		goCmdFlags       = goCmd.Flags()
		pluginCmdFlags   = pluginCmd.Flags()
		validateCmdFlags = validateCmd.Flags()
	)

	docCmdFlags.StringVar(&docgen.VarStringDir, "dir")
	docCmdFlags.StringVar(&docgen.VarStringOutput, "o")

	formatCmdFlags.StringVar(&format.VarStringDir, "dir")
	formatCmdFlags.BoolVar(&format.VarBoolIgnore, "iu")
	formatCmdFlags.BoolVar(&format.VarBoolUseStdin, "stdin")
	formatCmdFlags.BoolVar(&format.VarBoolSkipCheckDeclare, "declare")

	goCmdFlags.StringVar(&gogen.VarStringDir, "dir")
	goCmdFlags.StringVar(&gogen.VarStringAPI, "api")
	goCmdFlags.StringVar(&gogen.VarStringHome, "home")

	pluginCmdFlags.StringVarP(&plugin.VarStringPlugin, "plugin", "p")
	pluginCmdFlags.StringVar(&plugin.VarStringDir, "dir")
	pluginCmdFlags.StringVar(&plugin.VarStringAPI, "api")
	pluginCmdFlags.StringVar(&plugin.VarStringStyle, "style")

	validateCmdFlags.StringVar(&validate.VarStringAPI, "api")

	// Add sub-commands
	Cmd.AddCommand(docCmd, formatCmd, goCmd, pluginCmd, validateCmd)
}
