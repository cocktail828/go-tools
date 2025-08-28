package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

func assert(expr bool, format string, args ...any) {
	if !expr {
		panic(fmt.Sprintf(format, args...))
	}
}

var (
	reDirective = regexp.MustCompile(`^(\w+)(?:\s+([a-zA-Z0-9_]+(?:\s+[a-zA-Z0-9_]+)*))?`)
	reRoute     = regexp.MustCompile(`(\S+)\s+(\S+)(?:\s*\(([^\)]*)\))?(?:\s+return\s*(?:\(([^\)]*)\))?)?`)
)

func parserRoute(input string) ast.Route {
	match := reRoute.FindStringSubmatch(input)
	if match == nil {
		panic("invalid 'route' syntax, check your dsl file")
	}

	return ast.Route{
		Method:   strings.Title(match[1]),
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

	var dsl ast.DSL
	scanner := bufio.NewScanner(f)

	var curService *ast.Service
	var curGroup *ast.Group

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
					dsl.Services = append(dsl.Services, ast.Service{})
				} else {
					ss := strings.Fields(matches[2])
					for i, s := range ss {
						ss[i] = strings.Title(s)
					}

					dsl.Services = append(dsl.Services, ast.Service{Interceptors: ss})
				}
				curService = &dsl.Services[len(dsl.Services)-1]

			case "group":
				assert(len(matches) == 3, "invalid 'group' syntax, check your dsl file")

				args := strings.Fields(matches[2])
				if len(args) == 1 {
					curService.Groups = append(curService.Groups, ast.Group{Prefix: args[0]})
				} else {
					ss := args[1:]
					for i, s := range ss {
						ss[i] = strings.Title(s)
					}

					curService.Groups = append(curService.Groups, ast.Group{
						Prefix:       args[0],
						Interceptors: ss,
					})
				}
				curGroup = &curService.Groups[len(curService.Groups)-1]

			case "get", "head", "post", "put", "patch", "delete", "connect", "options", "trace":
				assert(len(matches) == 3, "invalid 'route' syntax, check your dsl file")
				curGroup.Routes = append(curGroup.Routes, parserRoute(line))

			default:
				return nil, fmt.Errorf("unknown keyword: %s", keyword)
			}
		}
	}

	if err := ast.Validate(dsl); err != nil {
		return nil, err
	}

	return &dsl, scanner.Err()
}
