package gen

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
)

type structureList struct {
	List StructureInfoList `json:"list"`
}

type structureFile struct {
	structureList `json:"structures"`
}

//go:embed schema.cue
var schemaData []byte

func readStructureInfoFIleSchema(ctx *cue.Context) (cue.Value, error) {
	value := ctx.CompileBytes(schemaData, cue.Filename("scheme.cue"))
	if value.Err() != nil {
		return cue.Value{}, value.Err()
	}
	return value, nil
}

func readYAML(ctx *cue.Context, r io.Reader) (cue.Value, error) {

	fileData, err := ioutil.ReadAll(r)
	if err != nil {
		return cue.Value{}, err
	}

	file, err := yaml.Extract("", fileData)

	if err != nil {
		return cue.Value{}, err
	}

	value := ctx.BuildFile(file, cue.Filename(""))
	if value.Err() != nil {
		return cue.Value{}, value.Err()
	}

	return value, nil
}

func ReadStructureInfoYAML(r io.Reader) (StructureInfoList, error) {
	ctx := cuecontext.New()

	schema, err := readStructureInfoFIleSchema(ctx)
	if err != nil {
		return nil, err
	}

	fileData, err := readYAML(ctx, r)
	if err != nil {
		return nil, err
	}

	fileDataValue := schema.Unify(fileData)
	if fileDataValue.Err() != nil {
		return nil, fmt.Errorf("ReadStructureYAML:  Invalid format")
	}

	var structureFileData structureFile
	err = fileDataValue.Decode(&structureFileData)
	if err != nil {
		return nil, err
	}

	if len(structureFileData.List) == 0 {
		return nil, errors.New("ReadStructureYAML: No structure data")
	}

	for name, structure := range structureFileData.List {
		if len(structure.Body) == 0 {
			return nil, fmt.Errorf("ReadStructureYAML: Name: %s No structure body", name)
		}
	}

	return structureFileData.List, nil
}
