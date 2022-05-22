package ueloghandler_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

type TestLogFile struct {
	fileName string
}

func NewTestLogFile() (*TestLogFile, error) {
	dirName, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(dirName, "Test.txt")
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	f.Close()

	return &TestLogFile{fileName: filename}, nil
}

func (f *TestLogFile) Close() {
	os.RemoveAll(filepath.Dir(f.fileName))
}

func (f *TestLogFile) Write(src io.Reader, bufferLen int, interval time.Duration) error {
	buffer := make([]byte, bufferLen)
	ticker := time.NewTicker(interval)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		f, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		_, err = f.Write(buffer[:n])
		if err != nil {
			return err
		}

		f.Close()
		if err != nil {
			return err
		}

		<-ticker.C
	}
}

func (f *TestLogFile) Name() string {
	return f.fileName
}

type FileNotifierTestCase struct {
	Name      string
	TestLog   string
	WantLogs  []string
	CRLFCheck bool
}

func (tc *FileNotifierTestCase) Run(t *testing.T) {

	assert := assert.New(t)
	tmpFile, err := NewTestLogFile()
	assert.NoError(err)
	defer tmpFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if tc.CRLFCheck {
		tc.TestLog = strings.NewReplacer(
			"\n", "\r\n",
		).Replace(tc.TestLog)
	}

	// write logs to temp log file for testing
	go func() {
		defer cancel()

		testLogReader := strings.NewReader(tc.TestLog)
		err = tmpFile.Write(testLogReader, 512, time.Millisecond*100)
		assert.NoError(err)
	}()

	notifier := ueloghandler.NewFileNotifier(tmpFile.Name(), time.Millisecond)

	receiveLogs := []string{}
	go func() {
		for log := range notifier.Logs() {
			receiveLogs = append(receiveLogs, log)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = notifier.Subscribe(ctx)
		assert.NoError(err)
	}()
	wg.Wait()

	notifier.Flush()

	// Wait receive log
	time.Sleep(time.Millisecond * 100)
	if assert.Equal(len(tc.WantLogs), len(receiveLogs)) {
		for i := 0; i < len(tc.WantLogs); i++ {
			assert.EqualValues(tc.WantLogs[i], receiveLogs[i])
		}
	}
}

func TestWatcher(t *testing.T) {
	testLog := `Log file open, 05/02/22 13:01:53
LogWindows: Failed to load 'aqProf.dll' (GetLastError=126)
LogInit: line1
line2
[2022.05.02-04.01.53:149][  0]LogConfig: CVar [[con.DebugLateDefault:1]] deferred - dummy variable created
[2022.05.02-04.01.58:862][970]LogHttp: Warning: warningline1
warningline2
warningline3
[2022.05.02-04.01.58:862][970]LogHttp: Error: errorline1
errorline2
errorline3
[2022.05.02-04.01.58:895][970]LogD3D11RHI: line1
line2
line3
[2022.05.02-14.10.33:382][513]LogTemp: Verbosity Log
[2022.05.02-14.10.33:382][513]LogTemp: Error: Verbosity Error
[2022.05.02-14.10.33:382][513]LogTemp: Warning: Verbosity Warning
[2022.05.02-14.10.33:382][513]LogTemp: Display: Verbosity Display
[2022.05.02-14.10.33:382][513]LogTemp: Verbose: Verbosity Verbose
[2022.05.02-14.10.33:382][513]LogTemp: VeryVerbose: Verbosity VeryVerbose
[2022.05.02-04.01.58:895][970]LogD3D11RHI: Shutdown
[2022.05.02-04.01.58:905][970]Log file closed, 05/02/22 13:01:58
`
	wantLog := []string{
		"Log file open, 05/02/22 13:01:53\n",
		"LogWindows: Failed to load 'aqProf.dll' (GetLastError=126)\n",
		"LogInit: line1\nline2\n",
		"[2022.05.02-04.01.53:149][  0]LogConfig: CVar [[con.DebugLateDefault:1]] deferred - dummy variable created\n",
		"[2022.05.02-04.01.58:862][970]LogHttp: Warning: warningline1\nwarningline2\nwarningline3\n",
		"[2022.05.02-04.01.58:862][970]LogHttp: Error: errorline1\nerrorline2\nerrorline3\n",
		"[2022.05.02-04.01.58:895][970]LogD3D11RHI: line1\nline2\nline3\n",
		"[2022.05.02-14.10.33:382][513]LogTemp: Verbosity Log\n",
		"[2022.05.02-14.10.33:382][513]LogTemp: Error: Verbosity Error\n",
		"[2022.05.02-14.10.33:382][513]LogTemp: Warning: Verbosity Warning\n",
		"[2022.05.02-14.10.33:382][513]LogTemp: Display: Verbosity Display\n",
		"[2022.05.02-14.10.33:382][513]LogTemp: Verbose: Verbosity Verbose\n",
		"[2022.05.02-14.10.33:382][513]LogTemp: VeryVerbose: Verbosity VeryVerbose\n",
		"[2022.05.02-04.01.58:895][970]LogD3D11RHI: Shutdown\n",
		"[2022.05.02-04.01.58:905][970]Log file closed, 05/02/22 13:01:58\n",
	}

	testCases := []FileNotifierTestCase{
		{
			Name:      "Basic",
			TestLog:   testLog,
			WantLogs:  wantLog,
			CRLFCheck: false,
		},
		{
			Name:      "CRLF",
			TestLog:   testLog,
			WantLogs:  wantLog,
			CRLFCheck: true,
		},
		{
			Name:      "BOM",
			TestLog:   "\ufeff" + testLog,
			WantLogs:  wantLog,
			CRLFCheck: false,
		},
		{
			Name:      "BOM_CRLF",
			TestLog:   "\ufeff" + testLog,
			WantLogs:  wantLog,
			CRLFCheck: true,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, testCase.Run)
	}
}
