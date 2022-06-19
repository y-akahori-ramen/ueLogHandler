package ueloghandler

import (
	"context"
	"sync"
)

type Watcher struct {
	handlerList []LogHandler
}

func NewWatcher() *Watcher {
	wacher := &Watcher{}
	return wacher
}

func (w *Watcher) AddLogHandler(handler LogHandler) {
	w.handlerList = append(w.handlerList, handler)
}

func (w *Watcher) Watch(ctx context.Context, notifier Notifier) error {
	eventHandleResult := make(chan error)

	var wg sync.WaitGroup
	watchEnd := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case logStr := <-notifier.Logs():
				log := NewLog(logStr)
				err := w.handleLog(log)
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
		err := notifier.Subscribe(ctx)
		if err == nil {
			err = notifier.Flush()
		}

		watchEnd <- struct{}{}
		eventHandleResult <- err
	}()

	err := <-eventHandleResult

	wg.Wait()
	return err
}

func (w *Watcher) handleLog(log Log) error {
	for _, handler := range w.handlerList {
		err := handler.HandleLog(log)
		if err != nil {
			return err
		}
	}
	return nil
}
