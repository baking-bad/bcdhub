package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

func newLogger() gl.Interface {
	writer := log.New(os.Stdout, "\r\n", log.LstdFlags)
	cfg := gl.Config{
		SlowThreshold: time.Second,
		LogLevel:      gl.Warn,
		Colorful:      true,
	}
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if cfg.Colorful {
		infoStr = gl.Green + "%s\n" + gl.Reset + gl.Green + "[info] " + gl.Reset
		warnStr = gl.BlueBold + "%s\n" + gl.Reset + gl.Magenta + "[warn] " + gl.Reset
		errStr = gl.Magenta + "%s\n" + gl.Reset + gl.Red + "[error] " + gl.Reset
		traceStr = gl.Green + "%s\n" + gl.Reset + gl.Yellow + "[%.3fms] " + gl.BlueBold + "[rows:%v]" + gl.Reset + " %s"
		traceWarnStr = gl.Green + "%s " + gl.Yellow + "%s\n" + gl.Reset + gl.RedBold + "[%.3fms] " + gl.Yellow + "[rows:%v]" + gl.Magenta + " %s" + gl.Reset
		traceErrStr = gl.RedBold + "%s " + gl.MagentaBold + "%s\n" + gl.Reset + gl.Yellow + "[%.3fms] " + gl.BlueBold + "[rows:%v]" + gl.Reset + " %s"
	}

	return &logger{
		Writer:       writer,
		Config:       cfg,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type logger struct {
	gl.Writer
	gl.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *logger) LogMode(level gl.LogLevel) gl.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gl.Info {
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gl.Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gl.Error {
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel > gl.Silent {
		elapsed := time.Since(begin)
		switch {
		case err != nil && l.LogLevel >= gl.Error:
			sql, rows := fc()
			if rows == -1 {
				l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gl.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case l.LogLevel == gl.Info:
			sql, rows := fc()
			if rows == -1 {
				l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}
