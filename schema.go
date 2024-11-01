package scheyaml

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/mitchellh/go-wordwrap"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

const nullValue = "null"

// scheYAML turns the schema into an example yaml tree, using fields such as default, description and examples.
//
//nolint:cyclop,gocognit // Perhaps in the future the code can be broken up in more pieces
func scheYAML(rootSchema *jsonschema.Schema, cfg *Config) *yaml.Node {
	result := new(yaml.Node)

	// This is to prevent a slice out of bounds panic, but shouldn't happen under normal circumstances
	if len(rootSchema.Type) == 0 {
		return result
	}

	// TODO: Currently we default to the first type in the list, if more types are defined they are ignored
	switch rootSchema.Type[0] {
	case "object":
		result.Kind = yaml.MappingNode
		properties := alphabeticalProperties(rootSchema)

		var requiredProperties []string
		for _, property := range properties {
			if slices.Contains(rootSchema.Required, property) {
				requiredProperties = append(requiredProperties, property)
			}
		}

		result.Content = []*yaml.Node{}
		for _, propertyName := range properties {
			property := (*rootSchema.Properties)[propertyName]
			overrideValue, hasOverrideValue := cfg.overrideFor(propertyName)
			if cfg.OnlyRequired && !hasOverrideValue && !slices.Contains(requiredProperties, propertyName) {
				continue
			}

			// Make sure that references are resolved on evaluation
			if property.Ref != "" {
				property = property.ResolvedRef
			}

			// The property name node
			result.Content = append(result.Content, &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       propertyName,
				HeadComment: formatHeadComment(property, cfg.LineLength),
			})

			if hasOverrideValue {
				// Otherwise it'd make it <nil>
				if overrideValue == nil {
					overrideValue = nullValue
				}

				// The property value node
				result.Content = append(result.Content, &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: fmt.Sprint(overrideValue),
				})

				continue
			}

			// The property value node
			valueNode := scheYAML(property, cfg.forProperty(propertyName))

			if valueNode.Content == nil && valueNode.Kind == yaml.MappingNode {
				valueNode.Value = "{}"
			}

			result.Content = append(result.Content, valueNode)
		}

	case "array":
		result.Kind = yaml.SequenceNode
		result.Content = []*yaml.Node{scheYAML(rootSchema.Items, cfg)}

	case nullValue:
		result.Kind = yaml.ScalarNode
		result.Value = nullValue

	// Leftover options: string, number, integer, boolean
	default:
		result.Kind = yaml.ScalarNode

		switch {
		case rootSchema.Default != nil:
			result.Value = fmt.Sprint(rootSchema.Default)

		default:
			result.LineComment = cfg.TODOComment
			result.Value = nullValue
		}
	}

	return result
}

// formatHeadComment will generate the comment above the property with the description
// and example values. The description will be word-wrapped in case it exceeds the given non-zero lineLength.
func formatHeadComment(schema *jsonschema.Schema, lineLength uint) string {
	var builder strings.Builder

	if schema.Description != nil {
		description := *schema.Description

		if lineLength > 0 {
			description = wordwrap.WrapString(*schema.Description, lineLength)
		}

		// Empty new lines aren't respected by default
		description = strings.ReplaceAll(description, "\n\n", "\n#\n")

		builder.WriteString(description)
	}

	if schema.Description != nil && len(schema.Examples) > 0 {
		// Empty newlines aren't respected, so we need to add our own #
		builder.WriteString("\n#\n")
	}

	if len(schema.Examples) > 0 {
		// Have too prepend a # here, newlines aren't commented by default
		builder.WriteString("Examples:\n")
		for _, example := range schema.Examples {
			_, _ = builder.WriteString("- ")
			if example != nil {
				_, _ = builder.WriteString(fmt.Sprint(example))
			} else {
				_, _ = builder.WriteString(nullValue)
			}
			_, _ = builder.WriteRune('\n')
		}
	}

	return builder.String()
}

// alphabeticalProperties is used to make the order of the object property deterministic. Might make this
// configurable later.
func alphabeticalProperties(schema *jsonschema.Schema) []string {
	result := maps.Keys(*schema.Properties)
	slices.Sort(result)
	return result
}
