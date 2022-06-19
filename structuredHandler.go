package ueloghandler

import (
	"encoding/json"
	"fmt"
)

type StructuredLogHandler struct {
	handlers map[string]StructuredLogDataHandler
}

func NewStructuredLogHandler() *StructuredLogHandler {
	return &StructuredLogHandler{handlers: make(map[string]StructuredLogDataHandler)}
}

func (h *StructuredLogHandler) AddHandler(handler StructuredLogDataHandler) error {
	if _, ok := h.handlers[handler.Type()]; ok {
		return fmt.Errorf("handler already exists: %s", handler.Type())
	} else {
		h.handlers[handler.Type()] = handler
		return nil
	}
}

func (h *StructuredLogHandler) HandleLog(log Log) error {
	type Header struct {
		Type string
	}
	type Body map[string]interface{}

	results, err := GetStructuredDataFromLog[Header, Body](log.Log)
	if err != nil {
		return err
	}

	if len(results) != 0 {
		for _, result := range results {
			structureType := result.Meta.Type
			if _, ok := h.handlers[structureType]; !ok {
				return fmt.Errorf("invalid structure type: %s", structureType)
			} else {
				jsonStr, err := json.Marshal(result.Body)
				if err != nil {
					return err
				}

				err = h.handlers[structureType].Handle(string(jsonStr), log)
				if err != nil {
					return err
				}
			}
		}
		return nil
	} else {
		return nil
	}
}

type StructuredLogDataHandler interface {
	Handle(json string, log Log) error
	Type() string
}

func NewStructuredLogDataHandler[THeader, TBody any](typeName string, handleFunc func(TStructuredData[THeader, TBody], Log) error) StructuredLogDataHandler {
	return &tStructuredLogDataHandler[THeader, TBody]{typeName: typeName, handleFunc: handleFunc}
}

type tStructuredLogDataHandler[THeader, TBody any] struct {
	handleFunc func(TStructuredData[THeader, TBody], Log) error
	typeName   string
}

func (h *tStructuredLogDataHandler[THeader, TBody]) Type() string {
	return h.typeName
}

func (h *tStructuredLogDataHandler[THeader, TBody]) Handle(json string, log Log) error {
	data, err := JSONToStructuredData[THeader, TBody](json)
	if err != nil {
		return err
	}

	return h.handleFunc(data, log)
}
