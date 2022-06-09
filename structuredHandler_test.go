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

type handlerTestHeader struct {
	HeaderString string
	HeaderInt    int
	HeaderFloat  float64
}

type handlerTestBody struct {
	BodyString string
	BodyInt    int
	BodyVector handlerTestVector
}

type handlerTestStructureData ueloghandler.TStructuredData[handlerTestHeader, handlerTestBody]

type handlerTestCase struct {
	structureTypeName string
	jsonStrings       []string
	wantErr           error
	wantData          []handlerTestStructureData
	actualData        []handlerTestStructureData
}

func (tc *handlerTestCase) Handle(data ueloghandler.TStructuredData[handlerTestHeader, handlerTestBody], log ueloghandler.Log) error {
	tc.actualData = append(tc.actualData, handlerTestStructureData(data))
	return nil
}

func TestStructuredLogHandler(t *testing.T) {
	testCases := []handlerTestCase{
		{
			structureTypeName: "TestStructure",
			jsonStrings: []string{
				`{"Header":{"HeaderString":"test1","HeaderInt":1,"HeaderFloat":1.0},"Body":{"BodyString":"test1","BodyInt":2,"BodyVector":{"X":1.1,"Y":1.2,"Z":1.3}}}`,
			},
			wantErr: nil,
			wantData: []handlerTestStructureData{
				{
					Meta: handlerTestHeader{
						HeaderString: "test1",
						HeaderInt:    1,
						HeaderFloat:  1.0,
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
				`{"Header":{"HeaderString":"testHeader1","HeaderInt":1,"HeaderFloat":1.0},"Body":{"BodyString":"testBody1","BodyInt":2,"BodyVector":{"X":1.1,"Y":1.2,"Z":1.3}}}`,
				`{"Header":{"HeaderString":"testHeader2","HeaderInt":2,"HeaderFloat":2.0},"Body":{"BodyString":"testBody2","BodyInt":3,"BodyVector":{"X":2.1,"Y":2.2,"Z":2.3}}}`,
			},
			wantErr: nil,
			wantData: []handlerTestStructureData{
				{
					Meta: handlerTestHeader{
						HeaderString: "testHeader1",
						HeaderInt:    1,
						HeaderFloat:  1.0,
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
					Meta: handlerTestHeader{
						HeaderString: "testHeader2",
						HeaderInt:    2,
						HeaderFloat:  2.0,
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
				`{"Header":{"HeaderString":"test1","HeaderInt":1,"HeaderFloat":1.0},"Body":{"BodyString":"test1","BodyInt":2,"BodyVector":{"X":1.1,"Y":1.2,"Z":1.3}}}`,
			},
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)

			logStr := ""
			for _, json := range testCase.jsonStrings {
				logStr += ueloghandler.BeginStructuredStr + fmt.Sprintf(`{"Header":{"Type":"%s"},"Body":%s}`, testCase.structureTypeName, json) + ueloghandler.EndStructuredStr
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
