package ueloghandler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

func TestNewLog(t *testing.T) {

	type testCase struct {
		name          string
		str           string
		wantCategory  string
		wantVerbosity string
		wantTime      string
		wantFrame     string
	}
	testCases := []testCase{
		{
			name:          "LogWithTime1",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Warning: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "Warning",
		},
		{
			name:          "LogWithTime2",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "",
		},
		{
			name: "LogWithTimeMultiline",
			str: `[2022.05.01-17.56.38:600][  9]LogTemp: Line1
Line2
	Line3`,
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogTemp",
			wantVerbosity: "",
		},
		{
			name:          "VerbosityInvalid",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Data: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "",
		},
		{
			name:          "VerbosityError",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Error: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "Error",
		},
		{
			name:          "VerbosityWarning",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Warning: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "Warning",
		},
		{
			name:          "VerbosityDisplay",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Display: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "Display",
		},
		{
			name:          "VerbosityVerbose",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: Verbose: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "Verbose",
		},
		{
			name:          "VerbosityVeryVerbose",
			str:           "[2022.05.01-17.56.38:600][  9]LogHttp: VeryVerbose: Cleaning up 0 outstanding Http requests.",
			wantTime:      "2022.05.01-17.56.38:600",
			wantFrame:     "  9",
			wantCategory:  "LogHttp",
			wantVerbosity: "VeryVerbose",
		},
		{
			name: "RawLog",
			str:  "Log file open, 05/02/22 13:01:53",
		},
		{
			name:         "CategoryOnly",
			str:          "LogWindows: Failed to load 'aqProf.dll' (GetLastError=126)",
			wantCategory: "LogWindows",
		},
		{
			name:         "PckagingLog",
			str:          "[2022.05.02-14.02.49:793][ 29]UATHelper: Packaging (Windows): LogShaderCompilers: Display:",
			wantTime:     "2022.05.02-14.02.49:793",
			wantCategory: "UATHelper",
			wantFrame:    " 29",
		},
		{
			name:         "Command not recognized log",
			str:          "[2022.05.22-01.11.05:634][733]Command not recognized: invalid command",
			wantTime:     "2022.05.22-01.11.05:634",
			wantCategory: "",
			wantFrame:    "733",
		},
		{
			name:         "Cmd log",
			str:          "[2022.05.22-01.11.05:633][733]Cmd: invalid command",
			wantTime:     "2022.05.22-01.11.05:633",
			wantCategory: "Cmd",
			wantFrame:    "733",
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			assert := assert.New(t)
			logInfo := ueloghandler.NewLog(testCase.str)
			wantLogInfo := ueloghandler.Log{
				Log:       testCase.str,
				Category:  testCase.wantCategory,
				Verbosity: testCase.wantVerbosity,
				Time:      testCase.wantTime,
				Frame:     testCase.wantFrame,
			}
			assert.Equal(wantLogInfo, logInfo)
		})
	}
}
