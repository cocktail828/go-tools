package docgen

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/parser"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var varStrDir, varStrOutput string
	cmd := &cobra.Command{
		Use:   "doc",
		Short: "Generate doc files in Markdown via API",
		Run: func(cmd *cobra.Command, args []string) {
			genDocCmd(varStrDir, varStrOutput)
		},
	}

	outputDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flags := cmd.Flags()
	flags.StringVar(&varStrDir, "dir", "", "")
	flags.StringVar(&varStrOutput, "o", outputDir, "")
	cmd.MarkFlagRequired("dir")

	return cmd
}

func genDocCmd(dir, outputDir string) {
	if _, err := os.Stat(dir); err != nil {
		log.Fatalf("dir %s not exsit", dir)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	var apifiles []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".api") {
			apifiles = append(apifiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range apifiles {
		api, err := parser.Parse(p, "")
		if err != nil {
			log.Fatalf("parse file: %s, err: %v", p, err)
		}

		api.Service = api.Service.JoinPrefix()
		err = genDoc(api, filepath.Dir(filepath.Join(outputDir, p[len(dir):])),
			strings.Replace(p[len(filepath.Dir(p)):], ".api", ".md", 1))
		if err != nil {
			log.Fatal(err)
		}
	}
}
