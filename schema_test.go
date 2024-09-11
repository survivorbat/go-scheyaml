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

func TestJSONSchemaObject_YAMLExample_ReturnsExpectedNodes(t *testing.T) {
	t.Parallel()
	// Arrange
	inputData, _ := os.ReadFile(path.Join("testdata", "test-schema.json"))

	var schemaObject jsonSchema
	err := json.Unmarshal(inputData, &schemaObject)
	require.NoError(t, err)

	// Act
	result := schemaObject.yamlExample()

	// Assert
	expectedData, _ := os.ReadFile(path.Join("testdata", "test-schema-result.yaml"))

	// Raw YAML from the node
	actualData, err := yaml.Marshal(&result)
	require.NoError(t, err)

	// First test the data itself, and quit if it isn't as expected.
	require.YAMLEq(t, string(expectedData), string(actualData))

	// If the properties are as expected, test the comments
	assert.Equal(t, string(expectedData), string(actualData))
}
