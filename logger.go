package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	gorm_logger "gorm.io/gorm/logger"
)

// newLogger returns a version of gorm's logger that shows loglines in a much more compact fashion.
func newLogger(writer gorm_logger.Writer, config gorm_logger.Config) gorm_logger.Interface {
	var (
		infoStr      = "%s [info] "
		warnStr      = "%s [warn] "
		errStr       = "%s [error] "
		traceStr     = "[%.0f ms] [%v] %s"
		traceWarnStr = "[%.0f ms] [%v] (%s) %s"
		traceErrStr  = "[%.0f ms] [%v] (%s) %s"
	)

	if config.Colorful {
		infoStr = gorm_logger.Green + "%s " + gorm_logger.Reset + gorm_logger.Green + "[info] " + gorm_logger.Reset
		warnStr = gorm_logger.BlueBold + "%s " + gorm_logger.Reset + gorm_logger.Magenta + "[warn] " + gorm_logger.Reset
		errStr = gorm_logger.Magenta + "%s " + gorm_logger.Reset + gorm_logger.Red + "[error] " + gorm_logger.Reset
		traceStr = gorm_logger.Yellow + "[%.0f ms] " + gorm_logger.BlueBold + "[%v]" + gorm_logger.Reset + " %s"
		traceWarnStr = gorm_logger.Yellow + "[%.0f ms] " + gorm_logger.BlueBold + "[%v]" + gorm_logger.RedBold + " %s" + gorm_logger.Reset + gorm_logger.Cyan + " %s" + gorm_logger.Reset
		traceErrStr = gorm_logger.Yellow + "[%.0f ms] " + gorm_logger.BlueBold + "[%v]" + gorm_logger.RedBold + " %s" + gorm_logger.Reset + "\n%s"
	}

	return &logger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type logger struct {
	gorm_logger.Writer
	gorm_logger.Config

	infoStr      string
	warnStr      string
	errStr       string
	traceStr     string
	traceErrStr  string
	traceWarnStr string
}

func (self *logger) LogMode(level gorm_logger.LogLevel) gorm_logger.Interface {
	newlogger := *self
	newlogger.LogLevel = level
	return &newlogger
}

func (self logger) Info(ctx context.Context, msg string, args ...any) {
	if self.LogLevel >= gorm_logger.Info {
		self.Printf(self.infoStr+msg, args...)
	}
}

func (self logger) Warn(ctx context.Context, msg string, args ...any) {
	if self.LogLevel >= gorm_logger.Warn {
		self.Printf(self.warnStr+msg, args...)
	}
}

func (self logger) Error(ctx context.Context, msg string, args ...any) {
	if self.LogLevel >= gorm_logger.Error {
		self.Printf(self.errStr+msg, args...)
	}
}

func (self logger) Trace(ctx context.Context, begin time.Time, f func() (string, int64), err error) {
	if self.LogLevel <= gorm_logger.Silent {
		return
	}

	r := func(rows int64) string {
		if rows == -1 {
			return "-"
		} else {
			return fmt.Sprint(rows)
		}
	}

	elapsed := time.Since(begin)
	ms := float64(elapsed) / float64(time.Millisecond)

	isSlow := elapsed > self.SlowThreshold && self.SlowThreshold != 0

	isErr := err != nil &&
		(!errors.Is(err, gorm_logger.ErrRecordNotFound) || !self.IgnoreRecordNotFoundError)

	switch {
	case isErr && self.LogLevel >= gorm_logger.Error:
		sql, rows := f()
		self.Printf(self.traceErrStr, ms, r(rows), err, sql)

	case isSlow && self.LogLevel >= gorm_logger.Warn:
		sql, rows := f()
		self.Printf(self.traceWarnStr, ms, r(rows), fmt.Sprintf(">= %v", self.SlowThreshold), sql)

	case self.LogLevel == gorm_logger.Info:
		sql, rows := f()
		self.Printf(self.traceStr, ms, r(rows), sql)
	}
}
