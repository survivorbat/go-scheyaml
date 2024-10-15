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
	// ID allows you to identify a schema, find more information here:
	// https://json-schema.org/understanding-json-schema/structuring#schema-identification
	//
	// Currently remains unused in scheyaml
	ID string `json:"$id,omitempty"`

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

	// Ref may be used to reference other parts of the JSON schema, find more information here:
	// https://json-schema.org/understanding-json-schema/structuring#schema-identification
	Ref string `json:"$ref,omitempty"`

	// misc contains all the leftover properties that we don't use, but want to preserve on a Unmarshal call
	misc map[string]any `json:"-"`
}

// ScheYAML turns the schema into an example yaml tree
//
//nolint:cyclop,gocognit // I don't think this method is worth breaking up yet
func (j *JSONSchema) ScheYAML(cfg *Config) (*yaml.Node, error) {
	// Make sure the first iteration this is set
	if cfg.rootSchema == nil {
		cfg.rootSchema = j
	}

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

			// Hydrate the schema if refs are used
			if err := property.ResolveRef(cfg); err != nil {
				return nil, fmt.Errorf("failed to resolve '%s': %w", property.Ref, err)
			}

			// The property name node
			result.Content = append(result.Content, &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       propertyName,
				HeadComment: property.formatHeadComment(cfg.LineLength),
			})

			if hasOverrideValue {
				// Otherwise it'd make it <nil>
				if overrideValue == nil {
					overrideValue = TypeNull.String()
				}

				result.Content = append(result.Content, &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: fmt.Sprint(overrideValue),
				})

				continue
			}

			// The property value node
			valueNode, err := property.ScheYAML(cfg.forProperty(propertyName))
			if err != nil {
				return nil, fmt.Errorf("failed to scheyaml property '%s': %w", propertyName, err)
			}

			if valueNode.Content == nil && valueNode.Kind == yaml.MappingNode {
				valueNode.Value = "{}"
			}

			result.Content = append(result.Content, valueNode)
		}

	case TypeArray:
		result.Kind = yaml.SequenceNode
		itemSchema, err := j.Items.ScheYAML(cfg)
		if err != nil {
			return nil, fmt.Errorf("could not scheyaml 'items': %w", err)
		}

		result.Content = []*yaml.Node{itemSchema}

	case TypeNull:
		result.Kind = yaml.ScalarNode
		result.Value = TypeNull.String()

	// Leftover options: string, number, integer, boolean
	default:
		result.Kind = yaml.ScalarNode

		switch {
		case j.Default != nil:
			result.Value = fmt.Sprint(j.Default)

		default:
			result.LineComment = cfg.TODOComment
			result.Value = TypeNull.String()
		}
	}

	return result, nil
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

// ResolveRef may be called to hydrate the current schema with its referenced properties. This
// is not done by default as we don't want to spam HTTP requests during the unmarshal step, however,
// calls such as ScheYAML will trigger these.
//
// These hydrated properties _will not_ be marshalled back into their schema counterpart,
// read the UnmarshalJSON docstring for more information.
func (j *JSONSchema) ResolveRef(cfg *Config) error {
	if j.Ref == "" {
		return nil
	}

	referencedSchema, err := lookupSchemaRef(cfg.rootSchema, j.Ref)
	if err != nil {
		return fmt.Errorf("failed to lookup '%s': %w", j.Ref, err)
	}

	j.ID = referencedSchema.ID
	j.Type = referencedSchema.Type
	j.Default = referencedSchema.Default
	j.Description = referencedSchema.Description
	j.Examples = referencedSchema.Examples
	j.Properties = referencedSchema.Properties
	j.Items = referencedSchema.Items
	j.Required = referencedSchema.Required

	// To allow deeper references, not for serialisation
	j.misc = referencedSchema.misc

	return nil
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

// MarshalJSON overrides the usual json behaviour to read data from the `misc` field and override
// whatever's in there with the available properties in the object.
//
// If a $ref is set on the object, all other properties (including misc) are ignored. This behaviour is described here:
// https://datatracker.ietf.org/doc/html/draft-pbryan-zyp-json-ref-03#section-3
func (j *JSONSchema) MarshalJSON() ([]byte, error) {
	if j.Ref != "" {
		return []byte(`{"$ref": "` + j.Ref + `"}`), nil
	}

	result := j.misc

	rawData, err := json.Marshal(safeJSONSchema(*j))
	if err != nil {
		return nil, err
	}

	// This can't fail
	_ = json.Unmarshal(rawData, &result)

	return json.Marshal(result)
}
