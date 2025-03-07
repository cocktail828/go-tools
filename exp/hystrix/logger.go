package hystrix

type logger interface {
	Printf(format string, items ...interface{})
}

// NoopLogger does not log anything.
type NoopLogger struct{}

// Printf does nothing.
func (l NoopLogger) Printf(format string, items ...interface{}) {}

// the default logger that will be used in the Hystrix package. By default prints nothing.
var log logger = NoopLogger{}

// SetLogger configures the logger that will be used. This only applies to the hystrix package.
func SetLogger(l logger) {
	log = l
}
