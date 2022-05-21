package ueloghandler

import (
	"errors"
	"strings"
	"time"
)

var ErrNoTimeData = errors.New("ueLogHandler:No time data")

type Log struct {
	Log          string
	Category     string
	Verbosity    string
	Time         string
	Frame        string
	FileOpenTime string
}

func (l *Log) ParseFileOpenTime(loc *time.Location) (time.Time, error) {
	if l.FileOpenTime == "" {
		return time.Time{}, ErrNoTimeData
	}

	const fileOpenTimeLayout = "01/02/06 15:04:05"
	return time.ParseInLocation(fileOpenTimeLayout, l.FileOpenTime, loc)
}

func (l *Log) ParseTime(loc *time.Location) (time.Time, error) {
	if l.Time == "" {
		return time.Time{}, ErrNoTimeData
	}
	const logTimeLayout = "2006.01.02-15.04.05.000"
	return time.ParseInLocation(logTimeLayout, strings.ReplaceAll(l.Time, ":", "."), loc)
}
