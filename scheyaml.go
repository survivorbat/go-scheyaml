package scheyaml

import (
	"fmt"

	"github.com/kaptinlin/jsonschema"
	"gopkg.in/yaml.v3"
)

// SchemaToYAML will take the given JSON schema and turn it into an example YAML file using fields like
// `description` and `examples` for documentation, `default` for default values and `properties` for listing blocks.
//
// You may provide options to customise the output.
func SchemaToYAML(schema *jsonschema.Schema, opts ...Option) ([]byte, error) {
	rootNode := SchemaToNode(schema, opts...)

	result, err := yaml.Marshal(&rootNode)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yaml nodes: %w", err)
	}

	return result, nil
}

// SchemaToNode is a lower-level version of SchemaToYAML, but returns the yaml.Node instead of the
// marshalled YAML.
//
// You may provide options to customise the output.
func SchemaToNode(schema *jsonschema.Schema, opts ...Option) *yaml.Node {
	config := NewConfig()

	for _, opt := range opts {
		opt(config)
	}

	return scheYAML(schema, config)
}
