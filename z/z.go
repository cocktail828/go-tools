package z

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/cocktail828/go-tools/z/reflectx"
)

// Abort: main.main
//   - Location:/root/github/go-tools/xlog/example/log.go:34 +0x48e649
//   - Detail: xxx 123
func Must(err error, args ...any) {
	if reflectx.IsNil(err) {
		return
	}

	funcname := ""
	file := "???"
	line := 0
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fs := runtime.CallersFrames([]uintptr{pc})
		f, _ := fs.Next()
		funcname = f.Function
		line = f.Line
		file = f.File
		if file == "" {
			file = "???"
		}
	}

	reason := "<No message provided>"
	if len(args) != 0 {
		reason = fmt.Sprintln(args...)
	}
	fmt.Printf("Abort: %s\n  - Location: %s:%d +0x%x\n  - Detail: %s\n",
		funcname, file, line, pc, reason)
	os.Exit(1)
}

func Mustf(err error, format string, args ...any) {
	if reflectx.IsNil(err) {
		return
	}

	funcname := ""
	file := "???"
	line := 0
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fs := runtime.CallersFrames([]uintptr{pc})
		f, _ := fs.Next()
		funcname = f.Function
		line = f.Line
		file = f.File
		if file == "" {
			file = "???"
		}
	}

	fmt.Printf("Abort: %s\n  - Location: %s:%d +0x%x\n  - Detail: %s\n",
		funcname, file, line, pc, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func DumpStack(depth, skip int) {
	pc := make([]uintptr, depth)
	n := runtime.Callers(skip+2, pc)
	pc = pc[:n]

	sb := strings.Builder{}
	frames := runtime.CallersFrames(pc)
	for i := 0; i < depth; i++ {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("Frame %d: %s\n\t%s:%d\n", i+1, frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	fmt.Print(sb.String())
}
