package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

func parseStructs(payload string) ([]ast.StructDef, error) {
	var structs []ast.StructDef

	structStartRegex := regexp.MustCompile(`struct\s+(\w+)\s*\{`)
	lines := strings.Split(payload, "\n")

	for i, line := range lines {
		startMatches := structStartRegex.FindStringSubmatch(line)
		if len(startMatches) < 2 {
			continue
		}

		structName := startMatches[1]

		endLineIdx := -1
		for j := i + 1; j < len(lines); j++ {
			if trimmedLine := strings.TrimSpace(lines[j]); trimmedLine == "}" {
				endLineIdx = j
				break
			}
		}

		if endLineIdx == -1 {
			return nil, fmt.Errorf("struct %s missing right '}'", structName)
		}

		fieldLines := lines[i+1 : endLineIdx]
		fields, err := parseStructFields(fieldLines)
		if err != nil {
			return nil, fmt.Errorf("parser struct %s field fail: %v", structName, err)
		}

		structs = append(structs, ast.StructDef{
			Name:   structName,
			Fields: fields,
		})
	}

	return structs, nil
}

func parseStructFields(fieldLines []string) ([]ast.StructField, error) {
	var fields []ast.StructField
	var inBlockComment bool         // 是否处于块注释 /* ... */ 中
	var blockCommentBuffer []string // 块注释内容缓存

	fieldRegex := regexp.MustCompile(
		`^\s*` + // 允许行首空格
			`(\w+)\s+` + // 字段名
			`(\w+)\s*` + // 类型
			`(` + "`[^`]*`" + `)?` + // 可选标签（`...`包裹）
			`\s*((//\s*(.*?))|(/\*.*?\*/)|(/\*.*?))?\s*$`, // 可选注释
	)

	for lineIdx, line := range fieldLines {
		lineNumber := lineIdx + 1
		trimmedLine := strings.TrimSpace(line)
		currentLine := line

		if trimmedLine == "" {
			continue
		}

		if inBlockComment {
			if strings.Contains(currentLine, "*/") {
				commentPart := strings.SplitN(currentLine, "*/", 2)[0]
				blockCommentBuffer = append(blockCommentBuffer, commentPart)
				inBlockComment = false

				currentLine = strings.SplitN(currentLine, "*/", 2)[1]
				trimmedLine = strings.TrimSpace(currentLine)
			} else {
				blockCommentBuffer = append(blockCommentBuffer, currentLine)
				continue
			}
		}

		for strings.Contains(currentLine, "/*") && !inBlockComment {
			parts := strings.SplitN(currentLine, "/*", 2)
			currentLine = parts[0]
			commentContent := parts[1]

			if strings.Contains(commentContent, "*/") {
				commentParts := strings.SplitN(commentContent, "*/", 2)
				blockCommentBuffer = append(blockCommentBuffer, commentParts[0])
				currentLine += commentParts[1]
			} else {
				blockCommentBuffer = append(blockCommentBuffer, commentContent)
				inBlockComment = true
				break
			}
		}

		if inBlockComment {
			continue
		}

		processedLine := currentLine
		trimmedProcessedLine := strings.TrimSpace(processedLine)
		if trimmedProcessedLine == "" || strings.HasPrefix(trimmedProcessedLine, "//") {
			if len(blockCommentBuffer) > 0 {
				blockCommentBuffer = []string{} // 清空缓存
			}
			continue
		}

		matches := fieldRegex.FindStringSubmatch(processedLine)
		if len(matches) < 3 {
			return nil, fmt.Errorf("第 %d 行字段格式无效: %s", lineNumber, line)
		}

		fieldName := matches[1]
		fieldType := matches[2]

		fieldTag := ""
		if len(matches) >= 4 && matches[3] != "" {
			fieldTag = strings.Trim(matches[3], "`")
		}

		var comments []string
		if len(blockCommentBuffer) > 0 {
			comments = append(comments, strings.Join(blockCommentBuffer, "\n"))
			blockCommentBuffer = []string{} // 清空缓存
		}

		if len(matches) >= 7 && matches[6] != "" {
			comments = append(comments, matches[6])
		}

		fieldComment := cleanCommentText(strings.Join(comments, "\n"))
		fields = append(fields, ast.StructField{
			Name:    fieldName,
			Type:    fieldType,
			Tag:     fieldTag,
			Comment: fieldComment,
		})
	}

	if inBlockComment {
		return nil, fmt.Errorf("存在未闭合的块注释 /*")
	}

	return fields, nil
}

// 清理注释文本，去除多余空格和星号
func cleanCommentText(comment string) string {
	lines := strings.Split(comment, "\n")
	var cleaned []string

	for _, line := range lines {
		// 去除行首的星号和空格
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "*")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, " ")
}
