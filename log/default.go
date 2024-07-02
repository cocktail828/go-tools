package log

import (
	"fmt"
	"io"
	"os"
)

var std = New(os.Stderr, "", LstdFlags)

// Default returns the standard logger used by the package-level output functions.
func Default() *Logger { return std }

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// Flags returns the output flags for the standard logger.
// The flag bits are Ldate, Ltime, and so on.
func Flags() int {
	return std.Flags()
}

// SetFlags sets the output flags for the standard logger.
// The flag bits are Ldate, Ltime, and so on.
func SetFlags(flag int) {
	std.SetFlags(flag)
}

// Prefix returns the output prefix for the standard logger.
func Prefix() string {
	return std.Prefix()
}

// SetPrefix sets the output prefix for the standard logger.
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}

// Writer returns the output destination for the standard logger.
func Writer() io.Writer {
	return std.Writer()
}

func Debugf(format string, v ...any) {
	std.levelOutput(0, 2, levelDebug, func(b []byte) []byte {
		return fmt.Appendf(b, format, v...)
	})
}

func Debug(v ...any) {
	std.levelOutput(0, 2, levelDebug, func(b []byte) []byte {
		return fmt.Appendln(b, v...)
	})
}

func Infof(format string, v ...any) {
	std.levelOutput(0, 2, levelInfo, func(b []byte) []byte {
		return fmt.Appendf(b, format, v...)
	})
}

func Info(v ...any) {
	std.levelOutput(0, 2, levelInfo, func(b []byte) []byte {
		return fmt.Appendln(b, v...)
	})
}

func Warnf(format string, v ...any) {
	std.levelOutput(0, 2, levelWarn, func(b []byte) []byte {
		return fmt.Appendf(b, format, v...)
	})
}

func Warn(v ...any) {
	std.levelOutput(0, 2, levelWarn, func(b []byte) []byte {
		return fmt.Appendln(b, v...)
	})
}

func Errorf(format string, v ...any) {
	std.levelOutput(0, 2, levelError, func(b []byte) []byte {
		return fmt.Appendf(b, format, v...)
	})
}

func Error(v ...any) {
	std.levelOutput(0, 2, levelError, func(b []byte) []byte {
		return fmt.Appendln(b, v...)
	})
}

// These functions write to the standard logger.

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...any) {
	std.output(0, 2, func(b []byte) []byte {
		return fmt.Append(b, v...)
	})
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...any) {
	std.output(0, 2, func(b []byte) []byte {
		return fmt.Appendf(b, format, v...)
	})
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...any) {
	std.output(0, 2, func(b []byte) []byte {
		return fmt.Appendln(b, v...)
	})
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...any) {
	std.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...any) {
	std.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...any) {
	std.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(v ...any) {
	s := fmt.Sprint(v...)
	std.Output(2, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	std.Output(2, s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	std.Output(2, s)
	panic(s)
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s) // +1 for this frame.
}
