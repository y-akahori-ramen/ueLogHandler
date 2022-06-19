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
      Meta:
        Tag: DataTag
        Insert: False
        Value: 1.23
        Value2: 123
      Body:
        Damage: int32
        Name: string
        Position: vector3`,
			expectData: map[string]gen.StructureInfo{
				"SampleStructure": {
					Meta: map[string]interface{}{
						"Tag":    "DataTag",
						"Insert": false,
						"Value":  1.23,
						"Value2": 123,
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
      Meta:
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
					Meta: map[string]interface{}{},
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
		{
			yaml: `
structures:
  list:
    sampleStructure:
      Body:
        Damage: int32`,
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
