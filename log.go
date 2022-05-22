package ueloghandler

import (
	"errors"
	"regexp"
	"strings"
	"time"
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

var timeFrameLogPattern = regexp.MustCompile(`^\[(\d{4}\.\d{2}\.\d{2}-\d{2}.\d{2}\.\d{2}:\d{3})\]\[((?:\s|\d){3})\](.+)`)
var categoryVerbosityPattern = regexp.MustCompile(`^([^:]+):\s((?:Error|Warning|Display|Verbose|VeryVerbose)):\s`)
var categoryPattern = regexp.MustCompile(`^([^:\s]+):\s(.+)`)

// NewLog Create log information from Unreal Engine format log
//
// Unreal engine log files contain multiple log formats.
// - Only log text
// - Log text with category
// - Log text with time, frame, and category
// Get log that match format whenever possible.
//
// Examples:
//  input: Log file open, 05/02/22 02:56:31
//	result: {Log:"Log file open, 05/02/22 02:56:31", Category:"", Verbosity:"", Time:"", Frame:""}
//
//	input: LogWindows: Failed to load 'aqProf.dll' (GetLastError=126)
//	result: {Log:"LogWindows: Failed to load 'aqProf.dll' (GetLastError=126)", Category:"LogWindows", Verbosity:"", Time:"", Frame:""}
//
//	input: [2022.05.01-17.56.38:615][429]LogTemp: Warning: WarningLog
//  result: Log:"[2022.05.01-17.56.38:615][429]LogTemp: Warning: WarningLog", Category:"LogTemp", Verbosity:"Warning", Time:"2022.05.01-17.56.38:615", Frame:"429"}
func NewLog(logStr string) Log {
	logInfo := Log{Log: logStr}

	var logWithoutTimeFrame string
	if timeFrameLogPattern.MatchString(logStr) {
		matches := timeFrameLogPattern.FindStringSubmatch(logStr)
		logInfo.Time = matches[1]
		logInfo.Frame = matches[2]
		logWithoutTimeFrame = matches[3]
	} else {
		logWithoutTimeFrame = logStr
	}

	if categoryVerbosityPattern.MatchString(logWithoutTimeFrame) {
		matches := categoryVerbosityPattern.FindStringSubmatch(logWithoutTimeFrame)
		logInfo.Category = matches[1]
		logInfo.Verbosity = matches[2]
	} else if categoryPattern.MatchString(logWithoutTimeFrame) {
		matches := categoryPattern.FindStringSubmatch(logWithoutTimeFrame)
		logInfo.Category = matches[1]
	}

	return logInfo
}

var convertUTF8_LFReplacer = strings.NewReplacer(
	"\r\n", "\n",
	"\ufeff", "",
)

func ToUTF8_LF(str string) string {
	return convertUTF8_LFReplacer.Replace(str)
}
