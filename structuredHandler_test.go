package ueloghandler_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

type handlerTestVector struct {
	X float64
	Y float64
	Z float64
}

type handlerTestMeta struct {
	MetaString string
	MetaInt    int
	MetaFloat  float64
}

type handlerTestBody struct {
	BodyString string
	BodyInt    int
	BodyVector handlerTestVector
}

type handlerTestStructureData ueloghandler.TStructuredData[handlerTestMeta, handlerTestBody]

type handlerTestCase struct {
	structureTypeName string
	jsonStrings       []string
	wantErr           error
	wantData          []handlerTestStructureData
	actualData        []handlerTestStructureData
}

func (tc *handlerTestCase) Handle(data ueloghandler.TStructuredData[handlerTestMeta, handlerTestBody], log ueloghandler.Log) error {
	tc.actualData = append(tc.actualData, handlerTestStructureData(data))
	return nil
}

func TestStructuredLogHandler(t *testing.T) {
	testCases := []handlerTestCase{
		{
			structureTypeName: "TestStructure",
			jsonStrings: []string{
				`{"Meta":{"MetaString":"test1","MetaInt":1,"MetaFloat":1.0},"Body":{"BodyString":"test1","BodyInt":2,"BodyVector":{"X":1.1,"Y":1.2,"Z":1.3}}}`,
			},
			wantErr: nil,
			wantData: []handlerTestStructureData{
				{
					Meta: handlerTestMeta{
						MetaString: "test1",
						MetaInt:    1,
						MetaFloat:  1.0,
					},
					Body: handlerTestBody{
						BodyString: "test1",
						BodyInt:    2,
						BodyVector: handlerTestVector{
							X: 1.1,
							Y: 1.2,
							Z: 1.3,
						},
					},
				},
			},
		},
		{
			structureTypeName: "TestStructure",
			jsonStrings: []string{
				`{"Meta":{"MetaString":"testMeta1","MetaInt":1,"MetaFloat":1.0},"Body":{"BodyString":"testBody1","BodyInt":2,"BodyVector":{"X":1.1,"Y":1.2,"Z":1.3}}}`,
				`{"Meta":{"MetaString":"testMeta2","MetaInt":2,"MetaFloat":2.0},"Body":{"BodyString":"testBody2","BodyInt":3,"BodyVector":{"X":2.1,"Y":2.2,"Z":2.3}}}`,
			},
			wantErr: nil,
			wantData: []handlerTestStructureData{
				{
					Meta: handlerTestMeta{
						MetaString: "testMeta1",
						MetaInt:    1,
						MetaFloat:  1.0,
					},
					Body: handlerTestBody{
						BodyString: "testBody1",
						BodyInt:    2,
						BodyVector: handlerTestVector{
							X: 1.1,
							Y: 1.2,
							Z: 1.3,
						},
					},
				},
				{
					Meta: handlerTestMeta{
						MetaString: "testMeta2",
						MetaInt:    2,
						MetaFloat:  2.0,
					},
					Body: handlerTestBody{
						BodyString: "testBody2",
						BodyInt:    3,
						BodyVector: handlerTestVector{
							X: 2.1,
							Y: 2.2,
							Z: 2.3,
						},
					},
				},
			},
		},
		{
			structureTypeName: "InvalidType",
			wantErr:           errors.New("invalid structure type: InvalidType"),
			jsonStrings: []string{
				`{"Meta":{"MetaString":"test1","MetaInt":1,"MetaFloat":1.0},"Body":{"BodyString":"test1","BodyInt":2,"BodyVector":{"X":1.1,"Y":1.2,"Z":1.3}}}`,
			},
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)

			logStr := ""
			for _, json := range testCase.jsonStrings {
				logStr += ueloghandler.BeginStructuredStr + fmt.Sprintf(`{"Meta":{"Type":"%s"},"Body":%s}`, testCase.structureTypeName, json) + ueloghandler.EndStructuredStr
			}

			logHandler := ueloghandler.NewStructuredLogHandler()

			dataHandler := ueloghandler.NewStructuredLogDataHandler("TestStructure", testCase.Handle)
			logHandler.AddHandler(dataHandler)

			err := logHandler.HandleLog(ueloghandler.Log{Log: logStr})
			assert.Equal(testCase.wantErr, err)
			assert.Equal(testCase.wantData, testCase.actualData)
		})
	}
}
