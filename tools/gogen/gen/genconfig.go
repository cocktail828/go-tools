package gen

import (
	_ "embed"
)

var (
	//go:embed genconfig.tpl
	genconfigTpl string
)

type GenConfig struct{}

func (g GenConfig) Gen(dsl *DSLMeta) (Writer, error) {
	return MultiFile{
		File{
			SubDir:  "config",
			Name:    "config.go",
			Payload: genconfigTpl,
		}, File{
			PlainTxt: true,
			Name:     "server.toml",
			Payload: `
[server]
addr = ":8080" # server addr
`,
		},
	}, nil
}
