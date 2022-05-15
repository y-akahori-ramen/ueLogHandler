package ueloghandler

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

var logFileOpenPattern = regexp.MustCompile(`Log\sfile\sopen,\s+(\S+\s+\S+)`)
var fileOpenAtTimeLayout = "01/02/06 15:04:05"

var ErrDetectFileOpenTime = errors.New("ueLogHandler:Fail to detect file open time")

type Log struct {
	unreallognotify.LogInfo
	FileOpenTime time.Time
}

type LogHandler func(Log) error

type Handler struct {
	Logs               chan Log
	handlerList        []LogHandler
	fileOpenAt         time.Time
	detectFileOpenTime bool
}

func NewHandler() *Handler {
	wacher := &Handler{Logs: make(chan Log)}
	return wacher
}

func (w *Handler) AddHandler(handler LogHandler) {
	w.handlerList = append(w.handlerList, handler)
}

func (w *Handler) Watch(ctx context.Context, filePath string, watchInterval time.Duration) error {
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
				if !w.detectFileOpenTime && log.Category == "" && logFileOpenPattern.MatchString(log.Log) {
					matches := logFileOpenPattern.FindStringSubmatch(log.Log)
					timeStr := matches[1]
					fileOpenAt, err := time.ParseInLocation(fileOpenAtTimeLayout, timeStr, time.Local)
					if err != nil {
						eventHandleResult <- err
						return
					}
					w.fileOpenAt = fileOpenAt
					w.detectFileOpenTime = true
				}

				if !w.detectFileOpenTime {
					eventHandleResult <- ErrDetectFileOpenTime
					return
				}

				logData := Log{LogInfo: log, FileOpenTime: w.fileOpenAt}
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

func (w *Handler) handleLog(log Log) error {
	for _, handler := range w.handlerList {
		err := handler(log)
		if err != nil {
			return err
		}
	}
	return nil
}
