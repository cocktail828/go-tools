package kits

import (
	"log"
	"reflect"
	"strings"
)

var (
	tree       = "├─"
	rightAngle = "└─"
	vertical   = "│ "
	blank      = "  "
)

// default logger
var Printf = log.Printf

func Tree(in any, prefixs ...string) {
	printStringer(in, false, prefixs)
}

// printStringer handles the recursive printing with indentation.
func printStringer(in any, istail bool, prefixs []string) {
	if in == nil {
		return
	}

	rt := reflect.TypeOf(in)
	rv := reflect.ValueOf(in)
	for rt.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return
		}
		rt = rt.Elem()
		rv = rv.Elem()
	}

	dumped := false
	if f, ok := in.(interface{ String() string }); ok {
		dumped = true
		Printf("%s%s", strings.Join(prefixs, ""), f.String())
	}

	switch rt.Kind() {
	case reflect.Struct:
		lastidx := lastSlice(rt, rv)
		for i := 0; i < rv.NumField(); i++ {
			if rt.Field(i).Tag.Get("dump") != "true" {
				continue
			}

			if fv := rv.Field(i); fv.CanInterface() {
				length := len(prefixs)
				switch {
				case length == 0:
				case prefixs[length-1] == tree:
					prefixs[length-1] = vertical
				case prefixs[length-1] == rightAngle:
					prefixs[length-1] = blank
				}
				if dumped {
					printStringer(fv.Interface(), lastidx == i, append(prefixs, blank))
				} else {
					printStringer(fv.Interface(), lastidx == i, prefixs)
				}
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			tailing := tree
			if i == rv.Len()-1 && istail {
				tailing = rightAngle
			}
			if len(prefixs) == 0 {
				tailing = blank
			}
			printStringer(rv.Index(i).Interface(), false, append(prefixs, tailing))
		}

	case reflect.String:
		Printf("%s%s", strings.Join(prefixs, ""), rv.Interface())
	}
}

func lastSlice(rt reflect.Type, rv reflect.Value) int {
	lastix := -1
	for i := 0; i < rv.NumField(); i++ {
		if rt.Field(i).Tag.Get("dump") != "true" {
			continue
		}

		if fv := rv.Field(i); fv.CanInterface() && fv.IsZero() {
			continue
		}
		lastix = i
	}
	return lastix
}
