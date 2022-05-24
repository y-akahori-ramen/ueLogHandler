package ueloghandler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

type TestNotifier struct {
	logChannel   chan string
	logs         []string
	sendInterval time.Duration
	curIdx       int
	returnValue  error
}

func NewTestNotifier(logs []string, sendInterval time.Duration, returnValue error) *TestNotifier {
	return &TestNotifier{
		logs:         logs,
		logChannel:   make(chan string),
		sendInterval: sendInterval,
		returnValue:  returnValue,
	}
}

func (t *TestNotifier) Logs() chan string {
	return t.logChannel
}

func (t *TestNotifier) Subscribe(ctx context.Context) error {
	if t.returnValue != nil {
		return t.returnValue
	}

	ticker := time.NewTicker(t.sendInterval)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if t.curIdx < len(t.logs) {
				t.logChannel <- t.logs[t.curIdx]
				t.curIdx++
			} else {
				return nil
			}
		}
	}
}

func (t *TestNotifier) Flush() error {
	for idx := t.curIdx; idx < len(t.logs); idx++ {
		t.logChannel <- t.logs[idx]
	}
	return nil
}

type TestNotifierTestCase struct {
	TestLog             []string
	WantLogs            []ueloghandler.Log
	WantWatcherLogs     []ueloghandler.WatcherLog
	WantErr             error
	WantFileOpenTime    []time.Time
	LogSendInterval     time.Duration
	WatcherTimeOut      time.Duration
	NotifierErr         error
	HandleErr           error
	receiveLogs         []ueloghandler.Log
	receiveWatcherLogs  []ueloghandler.WatcherLog
	receiveFileOpenTime []time.Time
}

func (tc *TestNotifierTestCase) Run(t *testing.T) {
	tc.receiveWatcherLogs = []ueloghandler.WatcherLog{}
	tc.receiveLogs = []ueloghandler.Log{}
	tc.receiveFileOpenTime = []time.Time{}

	watcher := ueloghandler.NewWatcher()
	watcher.AddWatcherLogHandler(ueloghandler.NewWatcherLogHandler(func(log ueloghandler.WatcherLog) error {
		if tc.HandleErr != nil {
			return tc.HandleErr
		}

		tc.receiveWatcherLogs = append(tc.receiveWatcherLogs, log)
		time, err := log.ParseFileOpenTime(time.UTC)
		if err != nil {
			return err
		}
		tc.receiveFileOpenTime = append(tc.receiveFileOpenTime, time)
		return nil
	}))
	watcher.AddLogHandler(ueloghandler.NewLogHandler(func(log ueloghandler.Log) error {
		if tc.HandleErr != nil {
			return tc.HandleErr
		}

		tc.receiveLogs = append(tc.receiveLogs, log)
		return nil
	}))

	notifier := NewTestNotifier(tc.TestLog, tc.LogSendInterval, tc.NotifierErr)
	ctx, cancel := context.WithTimeout(context.Background(), tc.WatcherTimeOut)
	defer cancel()
	err := watcher.Watch(ctx, notifier)

	assert := assert.New(t)
	assert.Equal(tc.WantErr, err)
	assert.Equal(tc.WantLogs, tc.receiveLogs)
	assert.Equal(tc.WantWatcherLogs, tc.receiveWatcherLogs)
	assert.Equal(tc.WantFileOpenTime, tc.receiveFileOpenTime)
}

func TestWatcher(t *testing.T) {
	testLogs := []string{
		"Log file open, 05/02/22 13:01:53\n",
		"[2022.05.02-04.01.58:905][970]Log file closed, 05/02/22 13:01:58\n",
		"Log file open, 05/02/23 13:01:53\n",
		"[2023.05.02-04.01.58:905][970]Log file closed, 05/02/22 13:01:58\n",
	}

	wantLogs := []ueloghandler.Log{
		{
			Log:       "Log file open, 05/02/22 13:01:53\n",
			Category:  "",
			Verbosity: "",
			Time:      "",
			Frame:     "",
		},
		{
			Log:       "[2022.05.02-04.01.58:905][970]Log file closed, 05/02/22 13:01:58\n",
			Category:  "",
			Verbosity: "",
			Time:      "2022.05.02-04.01.58:905",
			Frame:     "970",
		},
		{
			Log:       "Log file open, 05/02/23 13:01:53\n",
			Category:  "",
			Verbosity: "",
			Time:      "",
			Frame:     "",
		},
		{
			Log:       "[2023.05.02-04.01.58:905][970]Log file closed, 05/02/22 13:01:58\n",
			Category:  "",
			Verbosity: "",
			Time:      "2023.05.02-04.01.58:905",
			Frame:     "970",
		},
	}

	wantWatcherLogs := []ueloghandler.WatcherLog{
		{
			LogData:      wantLogs[0],
			FileOpenTime: "05/02/22 13:01:53",
		},
		{
			LogData:      wantLogs[1],
			FileOpenTime: "05/02/22 13:01:53",
		},
		{
			LogData:      wantLogs[2],
			FileOpenTime: "05/02/23 13:01:53",
		},
		{
			LogData:      wantLogs[3],
			FileOpenTime: "05/02/23 13:01:53",
		},
	}

	wantFileOpenTime := []time.Time{
		time.Date(2022, 5, 2, 13, 1, 53, 0, time.UTC),
		time.Date(2022, 5, 2, 13, 1, 53, 0, time.UTC),
		time.Date(2023, 5, 2, 13, 1, 53, 0, time.UTC),
		time.Date(2023, 5, 2, 13, 1, 53, 0, time.UTC),
	}

	testCases := []TestNotifierTestCase{
		{
			TestLog:          testLogs,
			WantLogs:         wantLogs,
			WantWatcherLogs:  wantWatcherLogs,
			LogSendInterval:  time.Millisecond,
			WatcherTimeOut:   time.Minute,
			WantFileOpenTime: wantFileOpenTime,
		},
		{
			TestLog:          testLogs,
			WantLogs:         wantLogs,
			WantWatcherLogs:  wantWatcherLogs,
			LogSendInterval:  time.Minute,
			WatcherTimeOut:   time.Millisecond,
			WantFileOpenTime: wantFileOpenTime,
		},
		{
			TestLog:          []string{},
			WantLogs:         []ueloghandler.Log{},
			WantWatcherLogs:  []ueloghandler.WatcherLog{},
			WantFileOpenTime: []time.Time{},
			LogSendInterval:  time.Millisecond,
			WatcherTimeOut:   time.Minute,
		},
		{
			TestLog:          []string{},
			WantLogs:         []ueloghandler.Log{},
			WantWatcherLogs:  []ueloghandler.WatcherLog{},
			WantFileOpenTime: []time.Time{},
			LogSendInterval:  time.Millisecond,
			WatcherTimeOut:   time.Minute,
			NotifierErr:      errors.New("Test"),
			WantErr:          errors.New("Test"),
		},
		{
			TestLog:          testLogs,
			WantLogs:         []ueloghandler.Log{},
			WantWatcherLogs:  []ueloghandler.WatcherLog{},
			WantFileOpenTime: []time.Time{},
			LogSendInterval:  time.Minute,
			WatcherTimeOut:   time.Millisecond,
			NotifierErr:      errors.New("Test"),
			WantErr:          errors.New("Test"),
		},
		{
			TestLog:          testLogs,
			WantLogs:         []ueloghandler.Log{},
			WantWatcherLogs:  []ueloghandler.WatcherLog{},
			WantFileOpenTime: []time.Time{},
			LogSendInterval:  time.Millisecond,
			WatcherTimeOut:   time.Minute,
			HandleErr:        errors.New("Test"),
			WantErr:          errors.New("Test"),
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), testCase.Run)
	}
}
