package scheyaml

import (
	"os"
	"path"
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestScheYAML_ReturnsExpectedNodesWithDefaults(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	cfg := NewConfig()

	// Act
	result, err := scheYAML(schema, cfg)

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

func TestScheYAML_ReturnsExpectedMinimalVersion(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-required.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.OnlyRequired = true

	// Act
	result, err := scheYAML(schema, cfg)

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

func TestScheYAML_ResolvesReferencesAtTheRootLevel(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData := `{
  "$ref": "#/$defs/MySchema",
  "$defs": {
    "MySchema": {
      "type": "object",
      "properties": {
        "name": {"type": "string", "default": "Hello World"}
      }
    }
  }
}`

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(inputData))
	require.NoError(t, err)

	cfg := NewConfig()

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)

	expectedData := "name: Hello World\n"

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, expectedData, string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, expectedData, string(actualData))
}

func TestScheYAML_ReturnsEmptyObjectOnNoProperties(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData := `{
  "type": "object",
  "properties": {
    "person": {
      "type": "object",
      "description": "Person"
    }
  }
}`

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(inputData))
	require.NoError(t, err)

	cfg := NewConfig()

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)

	expectedData := "# Person\nperson: {}\n"

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, expectedData, string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, expectedData, string(actualData))
}

// Catch-all for 'simple' overrides
func TestScheYAML_OverridesValuesFromConfig(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
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
	result, err := scheYAML(schema, cfg)

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
