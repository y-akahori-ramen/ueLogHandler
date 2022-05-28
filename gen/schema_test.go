package gen_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/y-akahori-ramen/ueLogHandler/gen"
)

func TestScheme(t *testing.T) {
	type testCase struct {
		yaml       string
		expectData gen.StructureInfoList
		noErr      bool
	}

	testCases := []testCase{
		{
			yaml: `
structures:
  list:
    SampleStructure:
      Header:
        Tag: DataTag
      Body:
        Damage: int32
        Name: string
        Position: vector3`,
			expectData: map[string]gen.StructureInfo{
				"SampleStructure": {
					Header: map[string]interface{}{
						"Tag": "DataTag",
					},
					Body: map[string]string{
						"Name":     "string",
						"Damage":   "int32",
						"Position": "vector3",
					},
				},
			},
			noErr: true,
		},
		{
			yaml: `
structures:
  list:
    SampleStructure:
      Header:
        Tag: DataTag`,
			expectData: nil,
			noErr:      false,
		},
		{
			yaml: `
structures:
  list:
    SampleStructure:
      Body:
        Damage: int32`,
			expectData: map[string]gen.StructureInfo{
				"SampleStructure": {
					Header: map[string]interface{}{},
					Body: map[string]string{
						"Damage": "int32",
					},
				},
			},
			noErr: true,
		},
		{
			yaml: `
structures:
  list:
    SampleStructure:
      Body:
        damage: int32`,
			expectData: nil,
			noErr:      false,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			assert := assert.New(t)

			readData, err := gen.ReadStructureInfoYAML(strings.NewReader(testCase.yaml))

			if testCase.noErr {
				assert.NoError(err)
			} else {
				assert.NotEqual(nil, err)
			}
			assert.Equal(testCase.expectData, readData)
		})
	}
}
