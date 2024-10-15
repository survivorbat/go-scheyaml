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

func TestJSONSchemaObject_ScheYAML_ReturnsExpectedNodesWithDefaults(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()

	// Act
	result, err := schemaObject.ScheYAML(cfg)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-output-defaults.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestJSONSchemaObject_ScheYAML_ReturnsExpectedMinimalVersion(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-required.json"))

	var schemaObject JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.OnlyRequired = true

	// Act
	result, err := schemaObject.ScheYAML(cfg)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-required-output.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

// Catch-all for 'simple' overrides
func TestJSONSchemaObject_ScheYAML_OverridesValuesFromConfig(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.TODOComment = "Perform magic here"
	cfg.LineLength = 0
	cfg.ValueOverrides = map[string]any{
		"numberProperty": 84,
		"stringProperty": nil, // Should be 'null' and not <nil>
		"objectProperty": map[string]any{
			"deepPropertyWithoutDescription": "b",
		},
	}

	// Act
	result, err := schemaObject.ScheYAML(cfg)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-output-overrides.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestJSONSchemaObject_ResolveRef_DoesNothingOnNoRef(t *testing.T) {
	t.Parallel()
	// Arrange
	input := `{
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "default": "Robin",
        "description": "The name of the customer"
      }
    }
  }`

	var schemaObject *JSONSchema
	err := json.Unmarshal([]byte(input), &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.rootSchema = schemaObject

	// Act
	err = schemaObject.Properties["name"].ResolveRef(cfg)

	// Assert
	require.NoError(t, err)

	nameProperty := schemaObject.Properties["name"]

	assert.Equal(t, TypeString, nameProperty.Type)
	assert.Equal(t, "Robin", nameProperty.Default)
	assert.Equal(t, "The name of the customer", nameProperty.Description)
}

func TestJSONSchemaObject_ResolveRef_ReturnsErrorOnNonExistingRef(t *testing.T) {
	t.Parallel()
	// Arrange
	input := `{
    "type": "object",
    "properties": {
      "name": {
        "$ref": "#/oops"
      }
    }
  }`

	var schemaObject *JSONSchema
	err := json.Unmarshal([]byte(input), &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.rootSchema = schemaObject

	// Act
	err = schemaObject.Properties["name"].ResolveRef(cfg)

	// Assert
	require.ErrorIs(t, err, ErrInvalidReference)
}

func TestJSONSchemaObject_ResolveRef_ResolvesExpectedReference(t *testing.T) {
	t.Parallel()
	// Arrange
	input := `{
    "type": "object",
    "properties": {
      "name": {
        "$ref": "#/$defs/NameSchema"
      }
    },
    "$defs": {
      "NameSchema": {
        "type": "string",
        "default": "Robin",
        "description": "The name of the customer",
        "examples": ["a", "b", "c"]
      }
    }
  }`

	var schemaObject *JSONSchema
	err := json.Unmarshal([]byte(input), &schemaObject)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.rootSchema = schemaObject

	// Act
	err = schemaObject.Properties["name"].ResolveRef(cfg)

	// Assert
	require.NoError(t, err)

	nameProperty := schemaObject.Properties["name"]

	assert.Equal(t, "#/$defs/NameSchema", nameProperty.Ref)

	assert.Equal(t, TypeString, nameProperty.Type)
	assert.Equal(t, "Robin", nameProperty.Default)
	assert.Equal(t, "The name of the customer", nameProperty.Description)
	assert.Equal(t, []any{"a", "b", "c"}, nameProperty.Examples)
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

func TestJSONSchemaObject_MarshalJSON_OnlyUnmarshalsRefProperty(t *testing.T) {
	t.Parallel()
	// Arrange
	input := &JSONSchema{
		Ref:         "#/$defs/HelloWorld",
		Description: "Hello World",
	}

	// Act
	result, err := json.Marshal(input)

	// Assert
	require.NoError(t, err)

	expected := `{"$ref": "#/$defs/HelloWorld"}`

	assert.JSONEq(t, expected, string(result))
}
