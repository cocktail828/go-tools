package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
	"github.com/pkg/errors"
)

func assert(expr bool, format string, args ...any) {
	if !expr {
		panic(fmt.Sprintf(format, args...))
	}
}

var (
	reDirective = regexp.MustCompile(`(@?[\w@.]+)(?:\s+([\w@./]+(?:\s+[\w@.]+)*))?`)
	reRoute     = regexp.MustCompile(`(\S+)\s+(\S+)(?:\s*\(([^\)]*)\))?(?:\s+return\s*(?:\(([^\)]*)\))?)?`)
)

func parserRoute(input string) ast.RouteAST {
	match := reRoute.FindStringSubmatch(input)
	if match == nil {
		panic("invalid 'route' syntax, check your dsl file")
	}

	return ast.RouteAST{
		Method:   match[1],
		Path:     match[2],
		Request:  match[3],
		Response: match[4],
	}
}

func ParseDSL(file string) (*ast.DSL, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	payload, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var dsl ast.DSL
	dsl.Structs, err = parseStructs(string(payload))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(payload))
	var curService *ast.ServiceAST
	var curGroup *ast.GroupAST
	var curHandlerNameLine string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if matches := reDirective.FindStringSubmatch(line); matches != nil {
			switch keyword := matches[1]; strings.ToLower(keyword) {
			case "syntax":
				assert(len(matches) == 3, "invalid 'syntax' syntax, check your dsl file")
				dsl.Syntax = matches[2]

			case "project":
				assert(len(matches) == 3, "invalid 'project' syntax, check your dsl file")
				dsl.Project = matches[2]

			case "service":
				assert(len(matches) >= 2, "invalid 'service' syntax, check your dsl file")
				assert(len(dsl.Services) == 0, "only 1 'service' should be defined")

				if len(matches) == 2 {
					dsl.Services = append(dsl.Services, ast.ServiceAST{})
				} else {
					dsl.Services = append(dsl.Services, ast.ServiceAST{Interceptors: strings.Fields(matches[2])})
				}
				curService = &dsl.Services[len(dsl.Services)-1]

			case "group":
				assert(len(matches) == 3, "invalid 'group' syntax, check your dsl file")

				args := strings.Fields(matches[2])
				if len(args) == 1 {
					curService.Groups = append(curService.Groups, ast.GroupAST{Prefix: args[0]})
				} else {
					curService.Groups = append(curService.Groups, ast.GroupAST{
						Prefix:       args[0],
						Interceptors: args[1:],
					})
				}
				curGroup = &curService.Groups[len(curService.Groups)-1]

			case "get", "head", "post", "put", "patch", "delete", "connect", "options", "trace":
				assert(len(matches) == 3, "invalid 'route' syntax, check your dsl file")
				rt := parserRoute(line)
				if curHandlerNameLine != "" {
					rt.HandlerName = curHandlerNameLine
					curHandlerNameLine = ""
				}

				if !slices.ContainsFunc(dsl.Structs, func(v ast.StructDef) bool {
					if rt.Request == "" || v.Name == rt.Request {
						return true
					}
					return false
				}) {
					return nil, errors.Errorf("request (%v) is not defined", rt.Request)
				}

				if !slices.ContainsFunc(dsl.Structs, func(v ast.StructDef) bool {
					if rt.Response == "" || v.Name == rt.Response {
						return true
					}
					return false
				}) {
					return nil, errors.Errorf("response (%v) is not defined", rt.Response)
				}

				curGroup.Routes = append(curGroup.Routes, rt)

			case "@handler":
				assert(len(matches) == 3, "invalid 'route.handler' syntax, check your dsl file")
				curHandlerNameLine = matches[2]

			default:
			}
		}
	}

	if err := ast.Validate(dsl); err != nil {
		return nil, err
	}

	return &dsl, scanner.Err()
}
