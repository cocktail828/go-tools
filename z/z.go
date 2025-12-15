package z

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/cocktail828/go-tools/z/reflectx"
)

func GoID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(idField, 10, 64)
	return id
}

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
	fmt.Printf("Abort: %s\n  - Location: %s:%d +0x%x\n  - Cause: %v\n  - Detail: %s\n",
		funcname, file, line, pc, err, reason)
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

	fmt.Printf("Abort: %s\n  - Location: %s:%d +0x%x\n  - Cause: %v\n  - Detail: %s\n",
		funcname, file, line, pc, err, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func Stack(depth int) string {
	pc := make([]uintptr, depth)
	n := runtime.Callers(2, pc)
	pc = pc[:n]

	if n == 0 {
		return ""
	}

	sb := strings.Builder{}
	frames := runtime.CallersFrames(pc)
	for i := range depth {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("Frame %d: %s\n\t%s:%d\n", i+1, frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}
