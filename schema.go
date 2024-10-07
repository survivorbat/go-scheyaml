package scheyaml

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

// Compile-time interface checks
var (
	_ json.Marshaler   = new(JSONSchema)
	_ json.Unmarshaler = new(JSONSchema)
)

// JSONSchema is used to unmarshal the incoming JSON schema into, does not support anyOf and allOf (yet).
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

	// Required properties set by parent object
	Required []string `json:"required,omitempty"`

	// misc contains all the leftover properties that we don't use, but want to preserve on a Unmarshal call
	misc map[string]any `json:"-"`
}

// ScheYAML turns the schema into an example yaml tree
//
//nolint:cyclop // I don't think this method is worth breaking up yet
func (j *JSONSchema) ScheYAML(cfg *Config) *yaml.Node {
	result := new(yaml.Node)

	//nolint:exhaustive // Not necessary, only array, null and object get special treatment
	switch j.Type {
	case TypeObject:
		result.Kind = yaml.MappingNode
		properties := j.alphabeticalProperties()

		var requiredProperties []string
		for _, property := range properties {
			if slices.Contains(j.Required, property) {
				requiredProperties = append(requiredProperties, property)
			}
		}

		result.Content = []*yaml.Node{}
		for _, propertyName := range properties {
			property := j.Properties[propertyName]
			overrideValue, hasOverrideValue := cfg.overrideFor(propertyName)
			if cfg.OnlyRequired && !hasOverrideValue && !slices.Contains(requiredProperties, propertyName) {
				continue
			}

			// The property name node
			result.Content = append(result.Content, &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       propertyName,
				HeadComment: property.formatHeadComment(cfg.LineLength),
			})

			if hasOverrideValue {
				result.Content = append(result.Content, &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: fmt.Sprint(overrideValue),
				})

				continue
			}

			// The property value node
			valueNode := property.ScheYAML(cfg.forProperty(propertyName))
			if valueNode.Content == nil && valueNode.Kind == yaml.MappingNode {
				valueNode.Value = "{}"
			}
			result.Content = append(result.Content, valueNode)
		}

	case TypeArray:
		result.Kind = yaml.SequenceNode
		result.Content = []*yaml.Node{j.Items.ScheYAML(cfg)}

	case TypeNull:
		result.Kind = yaml.ScalarNode
		result.Value = "null"

	// Leftover options: string, number, integer, boolean
	default:
		result.Kind = yaml.ScalarNode

		switch {
		case j.Default != nil:
			result.Value = fmt.Sprint(j.Default)

		default:
			result.LineComment = cfg.TODOComment
			result.Value = "null"
		}
	}

	return result
}

// formatHeadComment will generate the comment above the property with the description
// and example values. The description will be word-wrapped in case it exceeds the given non-zero lineLength.
func (j *JSONSchema) formatHeadComment(lineLength uint) string {
	var builder strings.Builder

	if j.Description != "" {
		description := j.Description

		if lineLength > 0 {
			description = wordwrap.WrapString(j.Description, lineLength)
		}

		// Empty new lines aren't respected by default
		description = strings.ReplaceAll(description, "\n\n", "\n#\n")

		builder.WriteString(description)
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
