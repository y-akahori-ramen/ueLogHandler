package gen

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

func toCppTypeName(typename string) (string, error) {
	switch typename {
	case "string":
		return "const FString&", nil
	case "vector2":
		return "const FVector2D&", nil
	case "vector3":
		return "const FVector&", nil
	case "float":
		fallthrough
	case "double":
		fallthrough
	case "int32":
		fallthrough
	case "uint32":
		fallthrough
	case "int64":
		fallthrough
	case "uint64":
		fallthrough
	case "bool":
		return typename, nil
	default:
		return "", fmt.Errorf("toCppTypeName:Invalid typename:%s", typename)
	}
}

func toCppTypeFormat(typename string) (string, error) {
	switch typename {
	case "string":
		return `"%s"`, nil
	case "vector2":
		return `{"X":%f,"Y":%f}`, nil
	case "vector3":
		return `{"X":%f,"Y":%f,"Z":%f}`, nil
	case "float":
		return "%f", nil
	case "double":
		return "%lf", nil
	case "int32":
		return "%d", nil
	case "uint32":
		return "%u", nil
	case "int64":
		return "%ld", nil
	case "uint64":
		return "%lu", nil
	case "bool":
		return `%s`, nil // true or false
	default:
		return "", fmt.Errorf("toCppTypeFormat:Invalid typename:%s", typename)
	}
}

func toCppFormatParam(typename, argname string) (string, error) {
	switch typename {
	case "string":
		return fmt.Sprintf("*%s", argname), nil
	case "vector2":
		return fmt.Sprintf("%s.X,%s.Y", argname, argname), nil
	case "vector3":
		return fmt.Sprintf("%s.X,%s.Y,%s.Z", argname, argname, argname), nil
	case "bool":
		return fmt.Sprintf(`%s?TEXT("true"):TEXT("false")`, argname), nil // true or false
	case "float":
		fallthrough
	case "double":
		fallthrough
	case "int32":
		fallthrough
	case "uint32":
		fallthrough
	case "int64":
		fallthrough
	case "uint64":
		return argname, nil
	default:
		return "", fmt.Errorf("toCppFormatParam:Invalid typename:%s", typename)
	}
}

type fieldInfo struct {
	FieldName, CppTypeName, CppFormat, CppParam string
}

func newFieldInfo(typeName, fieldName string) (fieldInfo, error) {

	cppFormat, err := toCppTypeFormat(typeName)
	if err != nil {
		return fieldInfo{}, err
	}

	cppParam, err := toCppFormatParam(typeName, fieldName)
	if err != nil {
		return fieldInfo{}, err
	}

	cppTypeName, err := toCppTypeName(typeName)
	if err != nil {
		return fieldInfo{}, err
	}

	return fieldInfo{FieldName: fieldName, CppTypeName: cppTypeName, CppFormat: cppFormat, CppParam: cppParam}, nil
}

type bodyField []fieldInfo

func (b bodyField) Generate(structureName string) (string, string, string, error) {
	var jsonStr strings.Builder
	var funcName strings.Builder
	var printParam strings.Builder

	_, err := funcName.WriteString(fmt.Sprintf("FString Log%s(", structureName))
	if err != nil {
		return "", "", "", err
	}

	_, err = jsonStr.WriteString("{")
	if err != nil {
		return "", "", "", err
	}

	for i, info := range b {
		_, err = jsonStr.WriteString(fmt.Sprintf(`"%s":`, info.FieldName))
		if err != nil {
			return "", "", "", err
		}

		_, err = jsonStr.WriteString(info.CppFormat)
		if err != nil {
			return "", "", "", err
		}

		_, err = funcName.WriteString(fmt.Sprintf("%s %s", info.CppTypeName, info.FieldName))
		if err != nil {
			return "", "", "", err
		}

		_, err := printParam.WriteString("," + info.CppParam)
		if err != nil {
			return "", "", "", err
		}

		if i < len(b)-1 {
			_, err = jsonStr.WriteString(",")
			if err != nil {
				return "", "", "", err
			}
			_, err = funcName.WriteString(",")
			if err != nil {
				return "", "", "", err
			}
		}
	}

	_, err = jsonStr.WriteString("}")
	if err != nil {
		return "", "", "", err
	}
	_, err = funcName.WriteString(")")
	if err != nil {
		return "", "", "", err
	}

	return jsonStr.String(), funcName.String(), printParam.String(), nil
}

func genCppLogCode(w io.Writer, structureName string, info StructureInfo) error {

	// Sorting to ensure key order
	bodyFieldNames := []string{}
	for fieldName := range info.Body {
		bodyFieldNames = append(bodyFieldNames, fieldName)
	}
	sort.Strings(bodyFieldNames)

	var bodyFieldData bodyField
	for _, fieldName := range bodyFieldNames {
		typeName := info.Body[fieldName]
		fieldInfo, err := newFieldInfo(typeName, fieldName)
		if err != nil {
			return err
		}
		bodyFieldData = append(bodyFieldData, fieldInfo)
	}

	bodyJson, funcName, printParam, err := bodyFieldData.Generate(structureName)
	if err != nil {
		return nil
	}

	bodyStrForMarshal := fmt.Sprintf("_REMOVEDOUBLEQUOTE_%s_REMOVEDOUBLEQUOTE_", bodyJson)
	printFormat := map[string]interface{}{
		"Meta": map[string]interface{}{
			"Type": structureName,
		},
		"Body": map[string]interface{}{
			"Meta": info.Meta,
			"Body": bodyStrForMarshal,
		},
	}

	jsonData, err := json.Marshal(printFormat)
	if err != nil {
		return err
	}

	// Remove escape sequence
	jsonStr := string(jsonData)
	jsonStr = strings.ReplaceAll(jsonStr, `\`, "")
	jsonStr = strings.ReplaceAll(jsonStr, `"_REMOVEDOUBLEQUOTE_`, "")
	jsonStr = strings.ReplaceAll(jsonStr, `_REMOVEDOUBLEQUOTE_"`, "")

	_, err = fmt.Fprintf(w, `%s
{
	return FString::Printf(TEXT(R"(%s%s%s)")%s);
}

`,
		funcName,
		ueloghandler.BeginStructuredStr, jsonStr, ueloghandler.EndStructuredStr, printParam,
	)

	return err
}

func GenCppFile(w io.Writer, namespace string, infoList StructureInfoList) error {
	fmt.Fprintln(w, "// Code generated by structuregen. DO NOT EDIT.")
	fmt.Fprintln(w, "#pragma once")
	fmt.Fprintln(w, `#include "CoreMinimal.h"`)

	if namespace != "" {
		_, err := fmt.Fprintf(w, "namespace %s {\n\n", namespace)
		if err != nil {
			return err
		}
	}

	for structureName, info := range infoList {
		err := genCppLogCode(w, structureName, info)
		if err != nil {
			return err
		}
	}

	if namespace != "" {
		_, err := fmt.Fprintln(w, "}")
		if err != nil {
			return err
		}
	}

	return nil
}
