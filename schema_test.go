package scheyaml

import (
	"os"
	"path"
	"regexp/syntax"
	"slices"
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

func TestScheYAML_AddsSchemaHeaderOnRequested(t *testing.T) {
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
	cfg.OutputHeader = "This is my favorite comment"

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)

	expectedData := "# This is my favorite comment\n# Person\nperson: {}\n"

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
		"stringProperty": NullValue,
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

func TestScheYAML_SkipsSentinelValue(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.ValueOverrides = map[string]any{
		"numberProperty": SkipValue,
	}

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)
	assert.False(t, slices.ContainsFunc(result.Content, func(node *yaml.Node) bool {
		return node.Value == "numberProperty"
	}))
}

func TestScheYAML_NestedPatternProperties(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	cfg := NewConfig()
	var overrides map[string]any
	overridesData, err := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties-overrides.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(overridesData, &overrides))
	cfg.ValueOverrides = overrides

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestScheYAML_PatternPropertiesInvalidRegex(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	delete(*schema.PatternProperties, "^.*$")
	(*schema.PatternProperties)["(.("] = &jsonschema.Schema{} // invalid
	cfg := NewConfig()

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	expected := &syntax.Error{}
	require.ErrorAs(t, err, &expected)
	require.ErrorContains(t, err, syntax.ErrMissingParen.String())
	require.Nil(t, result)
}

func TestScheYAML_MappingNodeOnlyRequired(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-non-required.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	cfg := NewConfig()
	cfg.OnlyRequired = true

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "config", result.Content[0].Value)
	assert.Equal(t, "{}", result.Content[1].Value)
}

func TestScheYAML_OverridesOnSchemaWithoutPropertiesDoesNotPanic(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-without-properties.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	cfg := NewConfig()
	var overrides map[string]any
	overridesData, err := os.ReadFile(path.Join("testdata", "test-schema-without-properties-overrides.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(overridesData, &overrides))
	cfg.ValueOverrides = overrides

	// Act
	result, err := scheYAML(schema, cfg)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-without-properties-overrides.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))
}

func TestScheYAML_NoType(t *testing.T) {
	t.Parallel()
	// Arrange
	schema := &jsonschema.Schema{}

	// Act
	node, err := scheYAML(schema, NewConfig())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, yaml.Node{}, *node)
}

func TestResolve_EmptySlice(t *testing.T) {
	t.Parallel()
	// Arrange
	schemas := []*jsonschema.Schema{}

	// Act
	resolved := resolve(schemas)

	// Assert
	assert.Empty(t, resolved)
}
