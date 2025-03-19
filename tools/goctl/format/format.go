package format

import (
	"bufio"
	stderr "errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/format/parser"
	apiF "github.com/cocktail828/go-tools/tools/goctl/internal/parser/format"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	leftParenthesis  = "("
	rightParenthesis = ")"
	leftBrace        = "{"
	rightBrace       = "}"
)

func Command() *cobra.Command {
	var (
		varStrDir               string // varStrDir describes the directory.
		varBoolSkipCheckDeclare bool   // varBoolSkipCheckDeclare describes whether to skip.
	)

	cmd := &cobra.Command{
		Use:   "format",
		Short: "Format goctl api files",
		Run: func(cmd *cobra.Command, args []string) {
			doFormat(varStrDir, varBoolSkipCheckDeclare)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&varStrDir, "dir", "", "")
	flags.BoolVar(&varBoolSkipCheckDeclare, "declare", false, "")
	cmd.MarkFlagRequired("dir")

	return cmd
}

// doFormat format api file
func doFormat(varStrDir string, varBoolSkipCheckDeclare bool) {
	errs := []error{}

	_, err := os.Lstat(varStrDir)
	if err != nil {
		log.Fatalf(varStrDir + ": No such file or directory")
	}

	filepath.Walk(varStrDir, func(path string, fi os.FileInfo, errBack error) (err error) {
		if strings.HasSuffix(path, ".api") {
			if err := ApiFormatByPath(path, varBoolSkipCheckDeclare); err != nil {
				errs = append(errs, errors.Wrap(err, fi.Name()))
			}
		}
		return nil
	})

	if ferr := stderr.Join(errs...); ferr != nil {
		log.Println(ferr)
	}
}

// apiFormatReader
// filename is needed when there are `import` literals.
func apiFormatReader(reader io.Reader, filename string, skipCheckDeclare bool) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	result, err := apiFormat(string(data), skipCheckDeclare, filename)
	if err != nil {
		return err
	}

	_, err = fmt.Print(result)
	return err
}

// ApiFormatByPath format api from file path
func ApiFormatByPath(apiFilePath string, skipCheckDeclare bool) error {
	return apiF.File(apiFilePath)
}

// removeComment filters comment content
func removeComment(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		return strings.TrimSpace(line[:commentIdx])
	}
	return strings.TrimSpace(line)
}

func apiFormat(data string, skipCheckDeclare bool, filename ...string) (string, error) {
	var err error
	if skipCheckDeclare {
		_, err = parser.ParseContentWithParserSkipCheckTypeDeclaration(data, filename...)
	} else {
		_, err = parser.ParseContent(data, filename...)
	}
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	s := bufio.NewScanner(strings.NewReader(data))
	tapCount := 0
	newLineCount := 0
	var preLine string
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 {
			if newLineCount > 0 {
				continue
			}
			newLineCount++
		} else {
			if preLine == rightBrace {
				builder.WriteString(pathx.NL)
			}
			newLineCount = 0
		}

		if tapCount == 0 {
			ft, err := formatGoTypeDef(line, s, &builder)
			if err != nil {
				return "", err
			}

			if ft {
				continue
			}
		}

		noCommentLine := removeComment(line)
		if noCommentLine == rightParenthesis || noCommentLine == rightBrace {
			tapCount--
		}
		if tapCount < 0 {
			line := strings.TrimSuffix(noCommentLine, rightBrace)
			line = strings.TrimSpace(line)
			if strings.HasSuffix(line, leftBrace) {
				tapCount++
			}
		}
		if line != "" {
			fmt.Fprintln(&builder, strings.Repeat("\t", tapCount))
		}
		builder.WriteString(line + pathx.NL)
		if strings.HasSuffix(noCommentLine, leftParenthesis) || strings.HasSuffix(noCommentLine, leftBrace) {
			tapCount++
		}
		preLine = line
	}

	return strings.TrimSpace(builder.String()), nil
}

func formatGoTypeDef(line string, scanner *bufio.Scanner, builder *strings.Builder) (bool, error) {
	noCommentLine := removeComment(line)
	tokenCount := 0
	if strings.HasPrefix(noCommentLine, "type") && (strings.HasSuffix(noCommentLine, leftParenthesis) ||
		strings.HasSuffix(noCommentLine, leftBrace)) {
		var typeBuilder strings.Builder
		typeBuilder.WriteString(mayInsertStructKeyword(line, &tokenCount) + pathx.NL)
		for scanner.Scan() {
			noCommentLine := removeComment(scanner.Text())
			typeBuilder.WriteString(mayInsertStructKeyword(scanner.Text(), &tokenCount) + pathx.NL)
			if noCommentLine == rightBrace || noCommentLine == rightParenthesis {
				tokenCount--
			}
			if tokenCount == 0 {
				ts, err := format.Source([]byte(typeBuilder.String()))
				if err != nil {
					return false, errors.New("error format \n" + typeBuilder.String())
				}

				result := strings.ReplaceAll(string(ts), " struct ", " ")
				result = strings.ReplaceAll(result, "type ()", "")
				builder.WriteString(result)
				break
			}
		}
		return true, nil
	}

	return false, nil
}

func mayInsertStructKeyword(line string, token *int) string {
	insertStruct := func() string {
		if strings.Contains(line, " struct") {
			return line
		}
		index := strings.Index(line, leftBrace)
		return line[:index] + " struct " + line[index:]
	}

	noCommentLine := removeComment(line)
	if strings.HasSuffix(noCommentLine, leftBrace) {
		*token++
		return insertStruct()
	}
	if strings.HasSuffix(noCommentLine, rightBrace) {
		noCommentLine = strings.TrimSuffix(noCommentLine, rightBrace)
		noCommentLine = removeComment(noCommentLine)
		if strings.HasSuffix(noCommentLine, leftBrace) {
			return insertStruct()
		}
	}
	if strings.HasSuffix(noCommentLine, leftParenthesis) {
		*token++
	}

	return line
}
