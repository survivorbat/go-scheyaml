package scheyaml

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupSchema_ReturnsExpectedSchema(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject *JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	// Act
	result, err := lookupSchemaRef(schemaObject, "#/$defs/HelloWorld")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, TypeString, result.Type)
	assert.Equal(t, "Hello World!", result.Default)
	assert.Equal(t, []any{"Hello", "World", "Foo"}, result.Examples)
	assert.Equal(t, "This property is for testing string Scalar nodes. On top of that, it will also check that this description wrapped into multiple new lines to keep it readable in the YAML output.\n\nAlso, native newlines in a description should be respected.", result.Description)
}

func TestLookupSchema_ReturnsErrorOnMissingSchema(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject *JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	// Act
	result, err := lookupSchemaRef(schemaObject, "#/$defs/Foo")

	// Assert
	require.ErrorIs(t, err, ErrInvalidReference)
	require.ErrorContains(t, err, "failed to lookup")

	assert.Nil(t, result)
}

func TestLookupSchema_ReturnsErrorOnMalformedSchema(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject *JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	// Not a schema, just foo
	schemaObject.misc = map[string]any{
		"$defs": map[string]any{"HelloWorld": "foo"},
	}

	// Act
	result, err := lookupSchemaRef(schemaObject, "#/$defs/HelloWorld")

	// Assert
	require.ErrorIs(t, err, ErrInvalidReference)
	require.ErrorContains(t, err, "failed to parse schema")

	assert.Nil(t, result)
}

func TestLookupSchema_ReturnsErrorOnURLReference(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject *JSONSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	// Act
	result, err := lookupSchemaRef(schemaObject, "https://example.com/my-schema")

	// Assert
	require.ErrorIs(t, err, ErrNotSupported)
	assert.Nil(t, result)
}

func TestLookup_ReturnsExpectedData(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    map[string]any
		expected any
	}{
		"#": {
			input:    map[string]any{"foo": "bar"},
			expected: map[string]any{"foo": "bar"},
		},
		"#/values": {
			input:    map[string]any{"values": []string{"foo", "bar"}},
			expected: []string{"foo", "bar"},
		},
		"#/definitions/foo": {
			input:    map[string]any{"definitions": map[string]any{"foo": "bar"}},
			expected: "bar",
		},
	}

	for inputReference, testData := range tests {
		t.Run(inputReference, func(t *testing.T) {
			t.Parallel()
			// Arrange
			segments := strings.Split(inputReference, "/")

			// Act
			result, err := lookup(testData.input, segments)

			// Assert
			require.NoError(t, err)

			assert.Equal(t, testData.expected, result)
		})
	}
}

func TestLookup_ReturnsErrorOnReferenceNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	object := map[string]any{"foo": map[string]any{"baz": "no"}}
	segments := []string{"#", "foo", "bar"}

	// Act
	result, err := lookup(object, segments)

	// Assert
	require.ErrorIs(t, err, ErrInvalidReference)
	require.ErrorContains(t, err, "failed to find 'bar'")

	assert.Nil(t, result)
}

// Different path in the code
func TestLookup_ReturnsErrorOnFinalSegmentNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	object := map[string]any{"foo": map[string]any{"bar": "no"}}
	segments := []string{"#", "boo", "bar"}

	// Act
	result, err := lookup(object, segments)

	// Assert
	require.ErrorIs(t, err, ErrInvalidReference)
	require.ErrorContains(t, err, "failed to find 'boo'")

	assert.Nil(t, result)
}

func TestLookup_ReturnsErrorOnReferenceNotADeepObject(t *testing.T) {
	t.Parallel()
	// Arrange
	object := map[string]any{"foo": "bar"}
	segments := []string{"#", "foo", "bar"}

	// Act
	result, err := lookup(object, segments)

	// Assert
	require.ErrorIs(t, err, ErrInvalidReference)
	require.ErrorContains(t, err, "segment 'foo' was not an object")

	assert.Nil(t, result)
}
