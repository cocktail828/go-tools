package logger

func NewNoopLogger() Logger {
	return noopLogger{}
}

var _ Logger = noopLogger{}

type noopLogger struct{}

func (l noopLogger) Debugw(msg string, args ...any) {}
func (l noopLogger) Debugf(msg string, args ...any) {}
func (l noopLogger) Infow(msg string, args ...any)  {}
func (l noopLogger) Infof(msg string, args ...any)  {}
func (l noopLogger) Warnw(msg string, args ...any)  {}
func (l noopLogger) Warnf(msg string, args ...any)  {}
func (l noopLogger) Errorw(msg string, args ...any) {}
func (l noopLogger) Errorf(msg string, args ...any) {}
func (l noopLogger) With(args ...any) Logger        { return l }
func (l noopLogger) WithGroup(name string) Logger   { return l }
