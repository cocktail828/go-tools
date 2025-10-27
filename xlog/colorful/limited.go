package colorful

import (
	"sync/atomic"

	"github.com/cocktail828/go-tools/xlog"
	"golang.org/x/time/rate"
)

type limitedColorful struct {
	*Logger
	limiter    *rate.Limiter
	supressMap map[xlog.Level]*atomic.Uint32 // how many log has been supressed...
}

func (l *limitedColorful) log(depth int, printer lvprinter, v ...any) {
	supressed := l.supressMap[printer.Level()]
	if l.limiter.Allow() {
		l.Logger.log(depth, printer, v...)
		if supressed.Load() > 0 {
			l.Logger.logf(depth, printer, "about %d lines of log has been supressed...", supressed.Load())
		}
		supressed.Store(0)
	} else {
		supressed.Add(1)
	}
}

func (l *limitedColorful) logln(depth int, printer lvprinter, v ...any) {
	if printer.Level() < l.level {
		return
	}

	supressed := l.supressMap[printer.Level()]
	if l.limiter.Allow() {
		l.Logger.logln(depth, printer, v...)
		if supressed.Load() > 0 {
			l.Logger.logf(depth, printer, "about %d lines of log has been supressed...", supressed.Load())
		}
		supressed.Store(0)
	} else {
		supressed.Add(1)
	}
}

func (l *limitedColorful) logf(depth int, printer lvprinter, format string, v ...any) {
	if printer.Level() < l.level {
		return
	}

	supressed := l.supressMap[printer.Level()]
	if l.limiter.Allow() {
		l.Logger.logf(depth, printer, format, v...)
		if supressed.Load() > 0 {
			l.Logger.logf(depth, printer, "about %d lines of log has been supressed...", supressed.Load())
		}
		supressed.Store(0)
	} else {
		supressed.Add(1)
	}
}

func (l *limitedColorful) Print(v ...any)                 { l.log(4, l.print, v...) }
func (l *limitedColorful) Println(v ...any)               { l.logln(4, l.print, v...) }
func (l *limitedColorful) Printf(format string, v ...any) { l.logf(4, l.print, format, v...) }

func (l *limitedColorful) Debug(v ...any)                 { l.log(4, l.debu, v...) }
func (l *limitedColorful) Debugln(v ...any)               { l.logln(4, l.debu, v...) }
func (l *limitedColorful) Debugf(format string, v ...any) { l.logf(4, l.debu, format, v...) }

func (l *limitedColorful) Info(v ...any)                 { l.log(4, l.info, v...) }
func (l *limitedColorful) Infoln(v ...any)               { l.logln(4, l.info, v...) }
func (l *limitedColorful) Infof(format string, v ...any) { l.logf(4, l.info, format, v...) }

func (l *limitedColorful) Warn(v ...any)                 { l.log(4, l.warn, v...) }
func (l *limitedColorful) Warnln(v ...any)               { l.logln(4, l.warn, v...) }
func (l *limitedColorful) Warnf(format string, v ...any) { l.logf(4, l.warn, format, v...) }

func (l *limitedColorful) Error(v ...any)                 { l.log(4, l.erro, v...) }
func (l *limitedColorful) Errorln(v ...any)               { l.logln(4, l.erro, v...) }
func (l *limitedColorful) Errorf(format string, v ...any) { l.logf(4, l.erro, format, v...) }
