package ueloghandler_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

func TestGetStructuredJson(t *testing.T) {
	type testCase struct {
		logStr string
		want   []string
	}
	testCases := []testCase{
		{
			logStr: "normal log",
			want:   []string{},
		},
		{
			logStr: `_BEGIN_STRUCTURED_{"Key":"Value1"}_END_STRUCTURED_`,
			want:   []string{`{"Key":"Value1"}`},
		},
		{
			logStr: `_BEGIN_STRUCTURED_{"Key":"Value1"}_END_STRUCTURED__BEGIN_STRUCTURED_{"Key":"Value2"}_END_STRUCTURED_`,
			want:   []string{`{"Key":"Value1"}`, `{"Key":"Value2"}`},
		},
		{
			logStr: `noramlog_BEGIN_STRUCTURED_{"Key":"Value1"}_END_STRUCTURED_normallog`,
			want:   []string{`{"Key":"Value1"}`},
		},
		{
			logStr: `normallog_BEGIN_STRUCTURED_{"Key":"Value1"}_END_STRUCTURED_normallog_BEGIN_STRUCTURED_{"Key":"Value2"}_END_STRUCTURED_normallog`,
			want:   []string{`{"Key":"Value1"}`, `{"Key":"Value2"}`},
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)
			result := ueloghandler.GetStructuredJsonFromLog(testCase.logStr)
			assert.Equal(testCase.want, result)
		})
	}
}

func TestToStructuredData(t *testing.T) {
	type Header struct {
		Type string
	}
	type TestStructuredData ueloghandler.TStructuredData[Header, map[string]interface{}]

	type testCase struct {
		jsonStr  string
		wantData TestStructuredData
		wantErr  error
	}
	testCases := []testCase{
		{
			jsonStr: `{"Header":{"Type":"TypeValue"},"Body":{"Sample":10,"Sample2":"Sample2Value"}}`,
			wantData: TestStructuredData{
				Meta: Header{Type: "TypeValue"},
				Body: map[string]interface{}{
					"Sample":  float64(10),
					"Sample2": "Sample2Value",
				},
			},
			wantErr: nil,
		},
		{
			jsonStr:  `{}`,
			wantData: TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			jsonStr:  `{"Header":{"Typee":"TypeValue"},"Date":{}}`,
			wantData: TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			jsonStr:  `{"Headere":{"Type":"TypeValue"},"Body":{"Sample":10,"Sample2":"Sample2Value"}}`,
			wantData: TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)
			result, err := ueloghandler.JSONToStructuredData[Header, map[string]interface{}](testCase.jsonStr)
			assert.Equal(testCase.wantErr, err)
			assert.Equal(testCase.wantData, TestStructuredData(result))
		})
	}
}

func TestToStructuredData2(t *testing.T) {
	type Header struct {
		Type string
	}
	type Vector struct {
		X float64
		Y float64
		Z float64
	}
	type Body struct {
		Name     string
		Position Vector
	}
	type TestStructuredData ueloghandler.TStructuredData[Header, Body]

	type testCase struct {
		jsonStr  string
		wantData TestStructuredData
		wantErr  error
	}
	testCases := []testCase{
		{
			jsonStr: `{"Header":{"Type":"TypeValue"},"Body":{"Name":"A","Position":{"X":0,"Y":1,"Z":2}}}`,
			wantData: TestStructuredData{
				Meta: Header{Type: "TypeValue"},
				Body: Body{
					Name:     "A",
					Position: Vector{X: 0, Y: 1, Z: 2},
				},
			},
			wantErr: nil,
		},
		{
			jsonStr:  `{"Header":{"Type":"TypeValue"},"Date":{"Name":"A","Position":{"X":0,"Y":1,"Z":2}}}`,
			wantData: TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			jsonStr:  `{"Header":{},"Body":{}}`,
			wantData: TestStructuredData{},
			wantErr:  nil,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)
			result, err := ueloghandler.JSONToStructuredData[Header, Body](testCase.jsonStr)
			assert.Equal(testCase.wantErr, err)
			assert.Equal(testCase.wantData, TestStructuredData(result))
		})
	}
}

func TestGetStructuredData(t *testing.T) {
	type Header struct {
		Type string
	}
	type TestStructuredData ueloghandler.TStructuredData[Header, map[string]interface{}]

	type testCase struct {
		logStr   string
		wantData []TestStructuredData
		wantErr  error
	}
	testCases := []testCase{
		{
			logStr: `_BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"Body":{"Sample":10,"Sample2":"Sample2Value"}}_END_STRUCTURED_`,
			wantData: []TestStructuredData{
				{
					Meta: Header{Type: "TypeValue"},
					Body: map[string]interface{}{
						"Sample":  float64(10),
						"Sample2": "Sample2Value",
					},
				},
			},
			wantErr: nil,
		},
		{
			logStr: `_BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"Body":{"Sample":{"X":1,"Y":2,"Z":3}}}_END_STRUCTURED_`,
			wantData: []TestStructuredData{
				{
					Meta: Header{Type: "TypeValue"},
					Body: map[string]interface{}{
						"Sample": map[string]interface{}{
							"X": float64(1),
							"Y": float64(2),
							"Z": float64(3),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			logStr:   `_BEGIN_STRUCTURED_{"InvalidHeader":{"Type":"TypeValue"},"Body":{"Sample":{"X":1,"Y":2,"Z":3}}}_END_STRUCTURED_`,
			wantData: []TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			logStr:   `_BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"InvalidData":{"Sample":{"X":1,"Y":2,"Z":3}}}_END_STRUCTURED_`,
			wantData: []TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			logStr:   `_BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"InvalidData":{"Sample":{"X":1,"Y":2,"Z":3}}}_END_STRUCTURED__BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"Body":{"Sample":10,"Sample2":"Sample2Value"}}_END_STRUCTURED_`,
			wantData: []TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			logStr:   `_BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"Body":{"Sample":10,"Sample2":"Sample2Value"}}_END_STRUCTURED__BEGIN_STRUCTURED_{"Header":{"Type":"TypeValue"},"InvalidData":{"Sample":{"X":1,"Y":2,"Z":3}}}_END_STRUCTURED_`,
			wantData: []TestStructuredData{},
			wantErr:  ueloghandler.ErrInvalidStructuredLogFormat,
		},
		{
			logStr:   `{"Header":{"Type":"TypeValue"},"InvalidData":{"Sample":{"X":1,"Y":2,"Z":3}}}`,
			wantData: []TestStructuredData{},
			wantErr:  nil,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)
			results, err := ueloghandler.GetStructuredDataFromLog[Header, map[string]interface{}](testCase.logStr)
			assert.Equal(testCase.wantErr, err)

			resultsTypeConvert := []TestStructuredData{}
			for _, result := range results {
				resultsTypeConvert = append(resultsTypeConvert, TestStructuredData(result))
			}
			assert.Equal(testCase.wantData, resultsTypeConvert)
		})
	}
}
