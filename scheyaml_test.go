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

func TestInvalidSchemaError_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	err := &InvalidSchemaError{
		Errors: map[string]*jsonschema.EvaluationError{
			"1": jsonschema.NewEvaluationError("keyword", "code", "message"),
			"2": jsonschema.NewEvaluationError("keyword", "code", "message"),
		},
	}

	// Act
	message := err.Error()

	// Assert
	assert.Equal(t, "1: message\n2: message\n", message)
}

func TestSchemaToYAML_ReturnsExpectedOutput(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	// Act
	result, err := SchemaToYAML(schema)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-output-defaults.yaml"))

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(result))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(result))
}

func TestSchemaToNode_ReturnsExpectedOutput(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	// Act
	result, err := SchemaToNode(schema)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-output-defaults.yaml"))

	actualData, err := yaml.Marshal(result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}

func TestSchemaToNode_InvalidSchema(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	var overrides map[string]any
	overridesData, err := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties-overrides.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(overridesData, &overrides))

	// Act
	result, err := SchemaToNode(schema, WithOverrideValues(overrides))

	// Assert
	var actual *InvalidSchemaError
	require.ErrorAs(t, err, &actual)
	assert.NotEmpty(t, actual.Errors)
	assert.Nil(t, result)
}

func TestSchemaToYAML_InvalidSchema(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties.json"))

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(inputData)
	require.NoError(t, err)

	var overrides map[string]any
	overridesData, err := os.ReadFile(path.Join("testdata", "test-schema-nested-pattern-properties-overrides.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(overridesData, &overrides))

	// Act
	result, err := SchemaToYAML(schema, WithOverrideValues(overrides))

	// Assert
	var actual *InvalidSchemaError
	require.ErrorAs(t, err, &actual)
	assert.NotEmpty(t, actual.Errors)
	assert.Nil(t, result)
}
