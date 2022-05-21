package ueloghandler

import (
	"context"
	"regexp"
	"sync"
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

var logFileOpenPattern = regexp.MustCompile(`Log\sfile\sopen,\s+(\S+\s+\S+)`)

type WatcherLog struct {
	LogData      Log
	FileOpenTime string
}

func (l *WatcherLog) ParseFileOpenTime(loc *time.Location) (time.Time, error) {
	if l.FileOpenTime == "" {
		return time.Time{}, ErrNoTimeData
	}

	const fileOpenTimeLayout = "01/02/06 15:04:05"
	return time.ParseInLocation(fileOpenTimeLayout, l.FileOpenTime, loc)
}

type WatcherLogHandler interface {
	HandleLog(log WatcherLog) error
}

func NewWatcherLogHandler(function func(log WatcherLog) error) WatcherLogHandler {
	return &funcWatcherLogHanlder{function: function}
}

type funcWatcherLogHanlder struct {
	function func(log WatcherLog) error
}

func (l *funcWatcherLogHanlder) HandleLog(log WatcherLog) error {
	return l.function(log)
}

type ignoreWatcherLogHandler struct {
	logHandler LogHandler
}

func (h *ignoreWatcherLogHandler) HandleLog(log WatcherLog) error {
	return h.logHandler.HandleLog(log.LogData)
}

type Watcher struct {
	handlerList  []WatcherLogHandler
	fileOpenTime string
}

func NewWatcher() *Watcher {
	wacher := &Watcher{}
	return wacher
}

func (w *Watcher) AddWatcherLogHandler(handler WatcherLogHandler) {
	w.handlerList = append(w.handlerList, handler)
}

func (w *Watcher) AddLogHandler(handler LogHandler) {
	w.handlerList = append(w.handlerList, &ignoreWatcherLogHandler{logHandler: handler})
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
				if log.Category == "" && logFileOpenPattern.MatchString(log.Log) {
					matches := logFileOpenPattern.FindStringSubmatch(log.Log)
					w.fileOpenTime = matches[1]
				}

				watcherLog := WatcherLog{
					LogData: Log{
						Log:       log.Log,
						Category:  log.Category,
						Verbosity: log.Verbosity,
						Time:      log.Time,
						Frame:     log.Frame,
					},
					FileOpenTime: w.fileOpenTime,
				}

				err := w.handleLog(watcherLog)
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

func (w *Watcher) handleLog(log WatcherLog) error {
	for _, handler := range w.handlerList {
		err := handler.HandleLog(log)
		if err != nil {
			return err
		}
	}
	return nil
}
