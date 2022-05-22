package ueloghandler

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"
)

var ErrFileRemoved = errors.New("ueLogHandler:File removed")

type FileNotifier struct {
	logs               chan string
	readBytes          int64
	sb                 strings.Builder
	basicFormatSection bool
	watchInterval      time.Duration
	filePath           string
}

func NewFileNotifier(filePath string, watchInterval time.Duration) *FileNotifier {
	wacher := &FileNotifier{logs: make(chan string), watchInterval: watchInterval, filePath: filePath}
	return wacher
}

func (f *FileNotifier) Flush() error {
	err := f.read(f.filePath)
	f.sendUnsentLog()
	return err
}

// sendUnsentLog
//
// The log is sent when the next log is started.
// Therefore, there may be an unsent log when the Watch method ends.
func (f *FileNotifier) sendUnsentLog() {
	if f.sb.Len() > 0 {
		f.logs <- f.sb.String()
		f.sb.Reset()
	}

	f.readBytes = 0
	f.sb.Reset()
	f.basicFormatSection = false
}

func (f *FileNotifier) Logs() chan string {
	return f.logs
}

// Subscribe Start watching log file and send logs to Logs channel
func (f *FileNotifier) Subscribe(ctx context.Context) error {
	f.sendUnsentLog()

	fstat, err := os.Stat(f.filePath)
	if err != nil {
		return err
	}
	latestModTime := fstat.ModTime()

	ticker := time.NewTicker(f.watchInterval)
	for {
		select {
		case <-ticker.C:
			fstat, err := os.Stat(f.filePath)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return ErrFileRemoved
				}
				return err
			}
			if !fstat.ModTime().Equal(latestModTime) {
				latestModTime = fstat.ModTime()
				err = f.read(f.filePath)
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (f *FileNotifier) read(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fs, err := file.Stat()
	if err != nil {
		return err
	}
	if fileRecreated := fs.Size() < f.readBytes; fileRecreated {
		f.sendUnsentLog()
	}

	_, err = file.Seek(f.readBytes, io.SeekStart)
	if err != nil {
		return err
	}

	bufreader := bufio.NewReader(file)
	for {
		lineData, err := bufreader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		f.readBytes += int64(len(lineData))

		lineStr := string(lineData)
		lineStr = ToUTF8_LF(lineStr)

		logInfo := NewLog(lineStr)

		startBasicFormatLog := logInfo.Category != "" || logInfo.Time != ""

		if f.basicFormatSection {
			// The log may output multiple lines.
			// Therefore, when the start of the next log output is detected, the string up to that point is as a log.
			if startBasicFormatLog {
				f.logs <- f.sb.String()
				f.sb.Reset()
				f.sb.WriteString(lineStr)
			} else {
				f.sb.WriteString(lineStr)
			}
		} else {
			if startBasicFormatLog {
				f.sb.Reset()
				f.sb.WriteString(lineStr)
				f.basicFormatSection = true
			} else {
				f.logs <- lineStr
			}
		}
	}
}
