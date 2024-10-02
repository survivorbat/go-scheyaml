package scheyaml

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestJSONSchemaObject_YAMLExample_ReturnsExpectedNodesWithDefaults(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()

	// Act
	result := schemaObject.ScheYAML(cfg)

	// Assert
	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-output-defaults.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestJSONSchemaObject_YAMLExample_ReturnsExpectedMinimalVersion(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-required.json"))

	var schemaObject JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.Minimal = true

	// Act
	result := schemaObject.ScheYAML(cfg)

	// Assert
	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-required-output.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestJSONSchemaObject_YAMLExample_OverridesValuesFromConfig(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.TODOComment = "Perform magic here"
	cfg.ValueOverrides = map[string]any{
		"numberProperty": 84,
		"objectProperty": map[string]any{
			"deepPropertyWithoutDescription": "b",
		},
	}

	// Act
	result := schemaObject.ScheYAML(cfg)

	// Assert
	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-output-overrides.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestJSONSchemaObject_JSON_PreservesExtraProperties(t *testing.T) {
	t.Parallel()
	// Arrange
	input := `{
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "default": "Robin",
        "description": "The name of the customer",
        "specialProperty": "abc"
      },
      "beverages": {
        "type": "array",
        "description": "A list of beverages the customer has consumed",
        "extraordinary property": 19,
        "items": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string", 
              "description": "The name of the beverage", 
              "examples": ["Coffee", "Tea", "Cappuccino"]
            },
            "price": {
              "type": "number",
              "description": "The price of the product",
              "default": 4.5
            }
          }
        }
      }
    }
  }`

	schemaObject := new(JSONSchema)

	// Act
	unmarshalErr := json.Unmarshal([]byte(input), &schemaObject)
	rawData, marshalErr := json.Marshal(schemaObject)

	// Assert
	require.NoError(t, unmarshalErr)
	require.NoError(t, marshalErr)

	assert.JSONEq(t, input, string(rawData))
}
