package assert

import (
	"fmt"
	"log"
	"os"
)

var (
	std = log.New(os.Stderr, "[DEBUG]{} ", log.LstdFlags)
)

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of [fmt.Print].
func Print(v ...any) {
	if !IsDebugMode() {
		return
	}

	std.Output(2, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of [fmt.Printf].
func Printf(format string, v ...any) {
	if !IsDebugMode() {
		return
	}

	std.Output(2, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of [fmt.Println].
func Println(v ...any) {
	if !IsDebugMode() {
		return
	}

	std.Output(2, fmt.Sprintln(v...))
}

// Fatal is equivalent to [Print] followed by a call to [os.Exit](1).
func Fatal(v ...any) {
	if !IsDebugMode() {
		return
	}

	std.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to [Printf] followed by a call to [os.Exit](1).
func Fatalf(format string, v ...any) {
	if !IsDebugMode() {
		return
	}

	std.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to [Println] followed by a call to [os.Exit](1).
func Fatalln(v ...any) {
	if !IsDebugMode() {
		return
	}

	std.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to [Print] followed by a call to panic().
func Panic(v ...any) {
	if !IsDebugMode() {
		return
	}

	s := fmt.Sprint(v...)
	std.Output(2, s)
	panic(s)
}

// Panicf is equivalent to [Printf] followed by a call to panic().
func Panicf(format string, v ...any) {
	if !IsDebugMode() {
		return
	}

	s := fmt.Sprintf(format, v...)
	std.Output(2, s)
	panic(s)
}

// Panicln is equivalent to [Println] followed by a call to panic().
func Panicln(v ...any) {
	if !IsDebugMode() {
		return
	}

	s := fmt.Sprintln(v...)
	std.Output(2, s)
	panic(s)
}
