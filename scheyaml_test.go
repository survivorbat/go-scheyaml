package scheyaml

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaToYAML_ReturnsExpectedOutput(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	// Act
	result, err := SchemaToYAML(inputData)

	// Assert
	require.NoError(t, err)

	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-result.yaml"))

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(result))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(result))
}
