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

type TStructuredData[TMeta, TBody any] struct {
	Meta TMeta
	Body TBody
}

type tStructuredDataForParse[TMeta, TBody any] struct {
	Meta *TMeta
	Body *TBody
}

func JSONToStructuredData[TMeta, TBody any](jsonStr string) (TStructuredData[TMeta, TBody], error) {
	var parsedData tStructuredDataForParse[TMeta, TBody]
	if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
		return TStructuredData[TMeta, TBody]{}, err
	}

	valid := parsedData.Meta != nil && parsedData.Body != nil
	if valid {
		return TStructuredData[TMeta, TBody]{Meta: *parsedData.Meta, Body: *parsedData.Body}, nil
	} else {
		return TStructuredData[TMeta, TBody]{}, ErrInvalidStructuredLogFormat
	}
}

func GetStructuredDataFromLog[TMeta, TBody any](logstr string) ([]TStructuredData[TMeta, TBody], error) {
	jsons := GetStructuredJsonFromLog(logstr)
	if len(jsons) == 0 {
		return []TStructuredData[TMeta, TBody]{}, nil
	}

	data := []TStructuredData[TMeta, TBody]{}
	for _, jsonStr := range jsons {
		newData, err := JSONToStructuredData[TMeta, TBody](jsonStr)
		if err != nil {
			return []TStructuredData[TMeta, TBody]{}, err
		} else {
			data = append(data, newData)
		}
	}

	return data, nil
}
