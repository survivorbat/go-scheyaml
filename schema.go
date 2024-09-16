package scheyaml

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

// Compile-time interface checks
var (
	_ json.Marshaler   = new(JSONSchema)
	_ json.Unmarshaler = new(JSONSchema)
)

// PropType is a limited set of options from: https://json-schema.org/understanding-json-schema/reference/type
type PropType string

const (
	TypeString  PropType = "string"
	TypeInteger PropType = "integer"
	TypeNumber  PropType = "number"
	TypeObject  PropType = "object"
	TypeArray   PropType = "array"
	TypeBoolean PropType = "boolean"
	TypeNull    PropType = "null"
)

// JSONSchema is used to unmarshal the incoming JSON schema into, does not support anyOf an allOf (yet).
// It only features properties used in the ScheYAML process. It does preserve additional properties in the
// json schema to unmarhsal/marshal without losing data.
//
// This object is part of the lower-level API, the package-level methods are for regular usage
type JSONSchema struct {
	// Type must always be set
	Type PropType `json:"type"`

	// Default, if set, will be the value pre-filled in the result
	Default any `json:"default,omitempty"`

	// Description will be put in a comment above the property in the result
	Description string `json:"description,omitempty"`

	// Examples will be put beneath the description in the comment
	Examples []any `json:"examples,omitempty"`

	// Properties is only used if type is `object`
	Properties map[string]*JSONSchema `json:"properties,omitempty"`

	// Items is only used if type is `array`
	Items *JSONSchema `json:"items,omitempty"`

	// misc contains all the leftover properties that we don't use, but want to preserve on a Unmarshal call
	misc map[string]any `json:"-"`
}

// yamlNodesPerField exists to prevent a magic number, it specifies that for property 'foo' in `foo: bar`
// 2 YAML nodes are required, one for 'foo' and one for 'bar'.
const yamlNodesPerField = 2

// ScheYAML turns the schema into an example yaml tree
func (j *JSONSchema) ScheYAML(cfg *Config) *yaml.Node {
	result := new(yaml.Node)

	switch j.Type {
	case TypeObject:
		result.Kind = yaml.MappingNode
		result.Content = make([]*yaml.Node, len(j.Properties)*yamlNodesPerField)

		properties := j.alphabeticalProperties()

		for index, propertyName := range properties {
			property := j.Properties[propertyName]

			// The property name node
			result.Content[index*yamlNodesPerField] = &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       propertyName,
				HeadComment: property.formatHeadComment(),
			}

			// Skip recursing further and override the value with whatever the config says
			if overrideValue, ok := cfg.overrideFor(propertyName); ok {
				// The property value node
				result.Content[index*yamlNodesPerField+1] = &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: fmt.Sprint(overrideValue),
				}

				continue
			}

			// The property value node
			result.Content[index*yamlNodesPerField+1] = property.ScheYAML(cfg.forProperty(propertyName))
		}

	case TypeArray:
		result.Kind = yaml.SequenceNode
		result.Content = []*yaml.Node{j.Items.ScheYAML(cfg)}

	case TypeNull:
		result.Kind = yaml.ScalarNode
		result.Value = "null"

	// Leftover options: string, number, integer, boolean, null
	default:
		result.Kind = yaml.ScalarNode

		if j.Default != nil {
			result.Value = fmt.Sprint(j.Default)
		} else {
			result.LineComment = cfg.TODOComment
		}
	}

	return result
}

// formatHeadComment will generate the comment above the property with the description
// and example values.
func (j *JSONSchema) formatHeadComment() string {
	var builder strings.Builder

	if j.Description != "" {
		builder.WriteString(j.Description)
	}

	if j.Description != "" && len(j.Examples) > 0 {
		// Empty newlines aren't respected, so we need to add our own #
		builder.WriteString("\n#\n")
	}

	if len(j.Examples) > 0 {
		// Have too prepend a # here, newlines aren't commented by default
		builder.WriteString("Examples:\n")
		for _, example := range j.Examples {
			_, _ = builder.WriteString("- ")
			if example != nil {
				_, _ = builder.WriteString(fmt.Sprint(example))
			} else {
				_, _ = builder.WriteString("null")
			}
			_, _ = builder.WriteRune('\n')
		}
	}

	return builder.String()
}

// alphabeticalProperties is used to make the order of the object property deterministic. Might make this
// configurable later.
func (j *JSONSchema) alphabeticalProperties() []string {
	result := maps.Keys(j.Properties)
	slices.Sort(result)
	return result
}

// safeJSONSchema does not inherit UnmarshalJSON or MarshalJSON, so it won't recursively call itself
type safeJSONSchema JSONSchema

// UnmarshalJSON overrides the usual json behaviour to preserve data in the `misc` field.
func (j *JSONSchema) UnmarshalJSON(input []byte) error {
	var result safeJSONSchema

	if err := json.Unmarshal(input, &result); err != nil {
		return err
	}

	if err := json.Unmarshal(input, &result.misc); err != nil {
		return err
	}

	*j = JSONSchema(result)

	return nil
}

// MarshalJSON overrides the usual json behaviour to read data from the `misc` field.
func (j *JSONSchema) MarshalJSON() ([]byte, error) {
	result := j.misc

	rawData, err := json.Marshal(safeJSONSchema(*j))
	if err != nil {
		return nil, err
	}

	// This can't fail
	_ = json.Unmarshal(rawData, &result)

	return json.Marshal(result)
}
