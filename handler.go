package ueloghandler

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

var logFileOpenPattern = regexp.MustCompile(`Log\sfile\sopen,\s+(\S+\s+\S+)`)
var fileOpenTimeLayout = "01/02/06 15:04:05"
var logTimeLayout = "2006.01.02-15.04.05.000"
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
	return time.ParseInLocation(fileOpenTimeLayout, l.FileOpenTime, loc)
}

func (l *Log) ParseTime(loc *time.Location) (time.Time, error) {
	if l.Time == "" {
		return time.Time{}, ErrNoTimeData
	}
	return time.ParseInLocation(logTimeLayout, strings.ReplaceAll(l.Time, ":", "."), loc)
}

type LogHandler func(Log) error

type Watcher struct {
	Logs         chan Log
	handlerList  []LogHandler
	fileOpenTime string
}

func NewWatcher() *Watcher {
	wacher := &Watcher{Logs: make(chan Log)}
	return wacher
}

func (w *Watcher) AddHandler(handler LogHandler) {
	w.handlerList = append(w.handlerList, handler)
}

func (w *Watcher) Watch(ctx context.Context, filePath string, watchInterval time.Duration) error {
	eventHandleResult := make(chan error)

	var wg sync.WaitGroup
	watchEnd := make(chan struct{})

	watcher := unreallognotify.NewWatcher(watchInterval)
	watcher.SetConvertUTF8LF(true)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case log := <-watcher.Logs:
				if w.fileOpenTime == "" && log.Category == "" && logFileOpenPattern.MatchString(log.Log) {
					matches := logFileOpenPattern.FindStringSubmatch(log.Log)
					w.fileOpenTime = matches[1]
				}

				logData := Log{
					Log:          log.Log,
					Category:     log.Category,
					Verbosity:    log.Verbosity,
					Time:         log.Time,
					Frame:        log.Frame,
					FileOpenTime: w.fileOpenTime,
				}

				err := w.handleLog(logData)
				if err != nil {
					eventHandleResult <- err
					return
				}
			case <-watchEnd:
				return
			}
		}
	}()

	go func() {
		err := watcher.Watch(ctx, filePath)
		watcher.Flush()
		watchEnd <- struct{}{}
		eventHandleResult <- err
	}()

	err := <-eventHandleResult

	wg.Wait()
	return err
}

func (w *Watcher) handleLog(log Log) error {
	for _, handler := range w.handlerList {
		err := handler(log)
		if err != nil {
			return err
		}
	}
	return nil
}
