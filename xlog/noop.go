package xlog

type NoopLogger struct{}

func (noop NoopLogger) Debugf(msg string, format string, args ...any) {}
func (noop NoopLogger) Debugln(msg string, args ...any)               {}
func (noop NoopLogger) Infof(msg string, format string, args ...any)  {}
func (noop NoopLogger) Infoln(msg string, args ...any)                {}
func (noop NoopLogger) Warnf(msg string, format string, args ...any)  {}
func (noop NoopLogger) Warnln(msg string, args ...any)                {}
func (noop NoopLogger) Errorf(msg string, format string, args ...any) {}
func (noop NoopLogger) Errorln(msg string, args ...any)               {}
func (noop NoopLogger) With(args ...any) Logger                       { return noop }
func (noop NoopLogger) WithGroup(name string) Logger                  { return noop }
