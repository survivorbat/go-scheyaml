package scheyaml

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Option is used to customise the SchemaToYAML output
type Option func(*config)

// config serves as the configuration object to allow customisation in SchemaToYAML
type config struct{}

// SchemaToYAML will take the given JSON schema and turn it into an example YAML file using fields like
// `description` and `examples` for documentation, `default` for default values and `properties` for listing blocks.
//
// There are currently no options available yet
func SchemaToYAML(schema []byte, _ ...Option) ([]byte, error) {
	var schemaObject jsonSchema
	if err := json.Unmarshal(schema, &schemaObject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema to jsonSchema object: %w", err)
	}

	rootNode := schemaObject.yamlExample()

	result, err := yaml.Marshal(&rootNode)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yaml nodes: %w", err)
	}

	return result, nil
}
