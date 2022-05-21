package ueloghandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

const BeginStructuredStr = "_BEGIN_STRUCTURED_"
const EndStructuredStr = "_END_STRUCTURED_"

var getStructuredJsonPattern = regexp.MustCompile(fmt.Sprintf(`^.*?%s(.+?)%s(.*)`, BeginStructuredStr, EndStructuredStr))

func GetStructuredJsonFromLog(logstr string) []string {
	jsonStr := []string{}

	targetStr := logstr
	for {
		matches := getStructuredJsonPattern.FindStringSubmatch(targetStr)
		if len(matches) == 0 {
			break
		}
		jsonStr = append(jsonStr, matches[1])
		targetStr = matches[2]
	}

	return jsonStr
}

var ErrInvalidStructuredLogFormat = errors.New("ueLogHandler:Invalid structured log format")

type TStructuredData[THeader, TBody any] struct {
	Header THeader
	Body   TBody
}

type tStructuredDataForParse[THeader, TBody any] struct {
	Header *THeader
	Body   *TBody
}

func JSONToStructuredData[THeader, TBody any](jsonStr string) (TStructuredData[THeader, TBody], error) {
	var parsedData tStructuredDataForParse[THeader, TBody]
	if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
		return TStructuredData[THeader, TBody]{}, err
	}

	valid := parsedData.Header != nil && parsedData.Body != nil
	if valid {
		return TStructuredData[THeader, TBody]{Header: *parsedData.Header, Body: *parsedData.Body}, nil
	} else {
		return TStructuredData[THeader, TBody]{}, ErrInvalidStructuredLogFormat
	}
}

func GetStructuredDataFromLog[THeader, TBody any](logstr string) ([]TStructuredData[THeader, TBody], error) {
	jsons := GetStructuredJsonFromLog(logstr)
	if len(jsons) == 0 {
		return []TStructuredData[THeader, TBody]{}, nil
	}

	data := []TStructuredData[THeader, TBody]{}
	for _, jsonStr := range jsons {
		newData, err := JSONToStructuredData[THeader, TBody](jsonStr)
		if err != nil {
			return []TStructuredData[THeader, TBody]{}, err
		} else {
			data = append(data, newData)
		}
	}

	return data, nil
}
