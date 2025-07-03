package scheyaml

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"gopkg.in/yaml.v3"
)

// ErrInvalidInput is returned if given parameters are invalid
var ErrInvalidInput = errors.New("invalid input given")

// InvalidSchemaError is returned when the schema is not valid, see jsonschema.Validate
type InvalidSchemaError struct {
	Errors map[string]*jsonschema.EvaluationError
}

// Error is a multiline string of the string->jsonschema.EvaluationError
func (e InvalidSchemaError) Error() string {
	var builder strings.Builder
	for _, key := range slices.Sorted(maps.Keys(e.Errors)) {
		builder.WriteString(fmt.Sprintf("%s: %s\n", key, e.Errors[key]))
	}

	return builder.String()
}

// SchemaToYAML will take the given JSON schema and turn it into an example YAML file using fields like
// `description` and `examples` for documentation, `default` for default values and `properties` for listing blocks.
//
// If any of the given properties match a regex given in pattern properties, the fields are added to the
// relevant fields.
//
// You may provide options to customise the output.
func SchemaToYAML(schema *jsonschema.Schema, opts ...Option) ([]byte, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil: %w", ErrInvalidInput)
	}

	rootNode, err := SchemaToNode(schema, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to scheyaml schema: %w", err)
	}

	config := NewConfig()
	for _, opt := range opts {
		opt(config)
	}

	writer := new(bytes.Buffer)

	encoder := yaml.NewEncoder(writer)
	if config.Indent != 0 {
		encoder.SetIndent(config.Indent)
	}

	encodeErr := encoder.Encode(&rootNode)
	if encodeErr != nil {
		return nil, fmt.Errorf("failed to marshal yaml nodes: %w", encodeErr)
	}

	return writer.Bytes(), nil
}

// SchemaToNode is a lower-level version of SchemaToYAML, but returns the yaml.Node instead of the
// marshalled YAML.
//
// You may provide options to customise the output.
func SchemaToNode(schema *jsonschema.Schema, opts ...Option) (*yaml.Node, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil: %w", ErrInvalidInput)
	}

	config := NewConfig()

	for _, opt := range opts {
		opt(config)
	}

	if !config.SkipValidate {
		res := schema.Validate(config.ValueOverrides)
		if errs := res.Errors; errs != nil {
			return nil, &InvalidSchemaError{Errors: errs}
		}
	}

	return scheYAML(schema, config)
}
