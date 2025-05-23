package hystrix

import "github.com/cocktail828/go-tools/xlog"

// the default logger that will be used in the Hystrix package. By default prints nothing.
var log xlog.Printer = xlog.NoopPrinter{}

// SetLogger configures the logger that will be used. This only applies to the hystrix package.
func SetLogger(l xlog.Printer) { log = l }
