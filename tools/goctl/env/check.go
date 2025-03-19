package env

import (
	"fmt"
	"strings"
	"time"

	"github.com/cocktail828/go-tools/tools/goctl/internal/env"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/spf13/cobra"
)

func check(_ *cobra.Command, _ []string) error {
	return Prepare(boolVarInstall, boolVarForce, boolVarVerbose)
}

func Prepare(install, force, verbose bool) error {
	pending := true
	colorful.Info("[goctl-env]: preparing to check env")
	defer func() {
		if p := recover(); p != nil {
			colorful.Errorf("%+v", p)
			return
		}
		if pending {
			colorful.Debug("\n[goctl-env]: congratulations! your goctl environment is ready!")
		} else {
			colorful.Error(`
[goctl-env]: check env finish, some dependencies is not found in PATH, you can execute
command 'goctl env check --install' to install it, for details, please execute command 
'goctl env check --help'`)
		}
	}()

	cmds := []env.Command{env.Protoc{}, env.ProtocGenGo{}, env.ProtocGenGoGrpc{}}
	for _, c := range cmds {
		time.Sleep(200 * time.Millisecond)
		if c.Exists() {
			colorful.Infof("[goctl-env]: %q is installed", c.Name())
			continue
		}

		colorful.Warnf("[goctl-env]: %q is not found in $PATH", c.Name())
		if install {
			doInstall := func() {
				colorful.Infof("[goctl-env]: preparing to install %q", c.Name())
				path, err := c.Install(env.Get(env.GoctlCache))
				if err != nil {
					colorful.Errorf("[goctl-env]: an error interrupted the installation: %+v", err)
					pending = false
				} else {
					colorful.Debugf("[goctl-env]: %q is already installed in %q", c.Name(), path)
				}
			}

			if force {
				doInstall()
				continue
			}
			colorful.Infof("[goctl-env]: do you want to install %q [y: YES, n: No]", c.Name())
			for {
				var in string
				fmt.Scanln(&in)
				var brk bool
				switch {
				case strings.EqualFold(in, "y"):
					doInstall()
					brk = true
				case strings.EqualFold(in, "n"):
					pending = false
					colorful.Infof("[goctl-env]: %q installation is ignored", c.Name())
					brk = true
				default:
					colorful.Error("[goctl-env]: invalid input, input 'y' for yes, 'n' for no")
				}
				if brk {
					break
				}
			}
		} else {
			pending = false
		}
	}
	return nil
}
