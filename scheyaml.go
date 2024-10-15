package scheyaml

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// SchemaToYAML will take the given JSON schema and turn it into an example YAML file using fields like
// `description` and `examples` for documentation, `default` for default values and `properties` for listing blocks.
//
// You may provide options to customise the output.
func SchemaToYAML(schema []byte, opts ...Option) ([]byte, error) {
	rootNode, err := SchemaToNode(schema, opts...)
	if err != nil {
		// Not wrapping the error because that was already done
		return nil, err
	}

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
func SchemaToNode(schema []byte, opts ...Option) (*yaml.Node, error) {
	config := NewConfig()

	for _, opt := range opts {
		opt(config)
	}

	var schemaObject JSONSchema
	if err := json.Unmarshal(schema, &schemaObject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema to jsonSchema object: %w", err)
	}

	return schemaObject.ScheYAML(config)
}
