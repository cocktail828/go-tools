package xlog

type NopLogger struct{}

func (nop NopLogger) Debugf(format string, args ...any) {}
func (nop NopLogger) Debugln(msg string, args ...any)   {}
func (nop NopLogger) Errorf(format string, args ...any) {}
func (nop NopLogger) Errorln(msg string, args ...any)   {}
func (nop NopLogger) Fatalf(format string, args ...any) {}
func (nop NopLogger) Fatalln(msg string, args ...any)   {}
func (nop NopLogger) Infof(format string, args ...any)  {}
func (nop NopLogger) Infoln(msg string, args ...any)    {}
func (nop NopLogger) Printf(format string, args ...any) {}
func (nop NopLogger) Println(msg string, args ...any)   {}
func (nop NopLogger) Warnf(format string, args ...any)  {}
func (nop NopLogger) Warnln(msg string, args ...any)    {}
func (nop NopLogger) With(args ...any) Logger           { return nop }
func (nop NopLogger) WithGroup(name string) Logger      { return nop }
