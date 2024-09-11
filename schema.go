package scheyaml

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

// jsonSchema is used to unmarshal the incoming JSON schema into, does not support anyOf an allOf (yet)
type jsonSchema struct {
	// Type should be one of:
	// - string
	// - number
	// - integer
	// - object
	// - array
	// - boolean
	// - null
	//
	// Source: https://json-schema.org/understanding-json-schema/reference/type
	Type string `json:"type"`

	// Default, if set, will be the value pre-filled in the result
	Default any `json:"default,omitempty"`

	// Description will be put in a comment above the property in the result
	Description string `json:"description,omitempty"`

	// Examples will be put beneath the description in the comment
	Examples []any `json:"examples,omitempty"`

	// Properties is only used if type is `object`
	Properties map[string]*jsonSchema `json:"properties,omitempty"`

	// Items is only used if type is `array`
	Items *jsonSchema `json:"items,omitempty"`
}

// yamlNodesPerField exists to prevent a magic number, it specifies that for property 'foo' in `foo: bar`
// 2 YAML nodes are required, one for 'foo' and one for 'bar'.
const yamlNodesPerField = 2

// yamlExample turns the schema into yaml nodes, which can then be marshalled later
func (j *jsonSchema) yamlExample() *yaml.Node {
	result := new(yaml.Node)

	switch j.Type {
	case "object":
		result.Kind = yaml.MappingNode
		result.Content = make([]*yaml.Node, len(j.Properties)*yamlNodesPerField)

		properties := j.alphabeticalProperties()

		for index, propertyName := range properties {
			property := j.Properties[propertyName]

			result.Content[index*yamlNodesPerField] = &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       propertyName,
				HeadComment: property.formatHeadComment(),
			}

			result.Content[index*yamlNodesPerField+1] = property.yamlExample()
		}

	case "array":
		result.Kind = yaml.SequenceNode
		result.Content = []*yaml.Node{j.Items.yamlExample()}

	case "null":
		result.Kind = yaml.ScalarNode
		result.Value = "null"

		// Leftover options: string, number, integer, boolean, null
	default:
		result.Kind = yaml.ScalarNode

		if j.Default != nil {
			result.Value = fmt.Sprint(j.Default)
		} else {
			result.LineComment = "TODO: Fill this in"
		}
	}

	return result
}

// formatHeadComment will generate the comment above the property with the description
// and example values.
func (j *jsonSchema) formatHeadComment() string {
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
func (j *jsonSchema) alphabeticalProperties() []string {
	result := maps.Keys(j.Properties)
	slices.Sort(result)
	return result
}
