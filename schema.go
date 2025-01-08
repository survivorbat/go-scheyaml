package scheyaml

import (
	"fmt"
	"regexp"
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
//nolint:cyclop // Slightly higher than allowed, but readable enough
func scheYAML(rootSchema *jsonschema.Schema, cfg *Config) (*yaml.Node, error) {
	result := new(yaml.Node)

	// If we're dealing with a reference, we'll continue with a resolved version of it
	if rootSchema.Ref != "" {
		return scheYAML(rootSchema.ResolvedRef, cfg)
	}

	// This is to prevent a slice out of bounds panic, but shouldn't happen under normal circumstances
	if len(rootSchema.Type) == 0 {
		return result, nil
	}

	// TODO: Currently we default to the first type in the list, if more types are defined they are ignored
	switch rootSchema.Type[0] {
	case "object":
		result.Kind = yaml.MappingNode
		objectContent, err := scheYAMLObject(rootSchema, cfg)
		if err != nil {
			return nil, err
		}

		result.Content = objectContent

	case "array":
		result.Kind = yaml.SequenceNode
		arrayContent, err := scheYAML(rootSchema.Items, cfg)
		if err != nil {
			return nil, err
		}

		result.Content = []*yaml.Node{arrayContent}

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

	return result, nil
}

// scheYAMLObject encapsulates the logic to scheYAML a schema of type "object"
//
//nolint:cyclop // Acceptable complexity, splitting this up is overkill
func scheYAMLObject(schema *jsonschema.Schema, cfg *Config) ([]*yaml.Node, error) {
	// If no properties were defined (somehow), return an empty object
	if schema.Properties == nil {
		return []*yaml.Node{{Kind: yaml.MappingNode, Value: "{}"}}, nil
	}

	properties := alphabeticalProperties(schema)

	var requiredProperties []string
	for _, property := range properties {
		if slices.Contains(schema.Required, property) {
			requiredProperties = append(requiredProperties, property)
		}
	}

	patternProperties, err := determinePatternProperties(schema, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to scheyaml pattern properties: %w", err)
	}

	//nolint:prealloc // We can't, false positive
	var result []*yaml.Node

	for _, propertyName := range properties {
		property := (*schema.Properties)[propertyName]
		overrideValue, hasOverrideValue := cfg.overrideFor(propertyName)
		if cfg.OnlyRequired && !hasOverrideValue && !slices.Contains(requiredProperties, propertyName) {
			continue
		}

		// Make sure that references are resolved on evaluation
		if property.Ref != "" {
			property = property.ResolvedRef
		}

		// The property name node
		result = append(result, &yaml.Node{
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
			result = append(result, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: fmt.Sprint(overrideValue),
			})

			continue
		}

		// The property value node
		valueNode, err := scheYAML(property, cfg.forProperty(propertyName))
		if err != nil {
			return nil, fmt.Errorf("failed to scheyaml %q: %w", propertyName, err)
		}

		if valueNode.Content == nil && valueNode.Kind == yaml.MappingNode {
			valueNode.Value = "{}"
		}

		if patternNodes, ok := patternProperties[propertyName]; ok {
			valueNode.Content = append(valueNode.Content, patternNodes...)
		}

		result = append(result, valueNode)
	}

	return result, nil
}

// determinePatternProperties's purpose is to generate additional nodes for properties that match
// defined patternProperties in the schema
func determinePatternProperties(schema *jsonschema.Schema, cfg *Config) (map[string][]*yaml.Node, error) {
	result := make(map[string][]*yaml.Node)

	if schema.Properties == nil || schema.PatternProperties == nil {
		return result, nil
	}

	properties := maps.Keys(*schema.Properties)

	for regex, patternProperty := range *schema.PatternProperties {
		parsedRegex, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %q as a regex: %w", regex, err)
		}

		for _, property := range properties {
			if !parsedRegex.MatchString(property) {
				continue
			}

			result[property], err = scheYAMLObject(patternProperty, cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to scheyaml %q: %w", regex, err)
			}
		}
	}

	return result, nil
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
