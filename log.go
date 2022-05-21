package ueloghandler

import (
	"errors"
	"strings"
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

var ErrNoTimeData = errors.New("ueLogHandler:No time data")

type Log struct {
	Log       string
	Category  string
	Verbosity string
	Time      string
	Frame     string
}

func (l *Log) ParseTime(loc *time.Location) (time.Time, error) {
	if l.Time == "" {
		return time.Time{}, ErrNoTimeData
	}
	const logTimeLayout = "2006.01.02-15.04.05.000"
	return time.ParseInLocation(logTimeLayout, strings.ReplaceAll(l.Time, ":", "."), loc)
}

func NewLog(logStr string) Log {
	log := unreallognotify.NewLogInfo(logStr)
	return Log{
		Log:       log.Log,
		Category:  log.Category,
		Verbosity: log.Verbosity,
		Time:      log.Time,
		Frame:     log.Frame,
	}
}
