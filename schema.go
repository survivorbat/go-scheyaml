package scheyaml

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/mitchellh/go-wordwrap"
	"gopkg.in/yaml.v3"
)

// ErrParsing is returned if during evaluation of the schema an error occurs
var ErrParsing = errors.New("failed to parse/process jsonschema")

// NullValue when setting an override to the 'null' value. This can be used as opposed to
// SkipValue where the key is omitted entirely
const NullValue = "null"

// skipValue type alias used as a sentinel to omit a particular key from the result
type skipValue bool

// SkipValue can be set on any key to signal that it should be emitted from the result set
var SkipValue skipValue = true

// scheYAML turns the schema into an example yaml tree, using fields such as default, description and examples.
func scheYAML(rootSchema *jsonschema.Schema, cfg *Config) (*yaml.Node, error) { //nolint:cyclop // accepted complexity
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
		if len(cfg.ItemsOverrides) > 0 {
			for i := range len(cfg.ItemsOverrides) {
				arrayContent, err := scheYAML(rootSchema.Items, cfg.forIndex(i))
				if err != nil {
					return nil, err
				}

				result.Content = append(result.Content, arrayContent)
			}

			break
		}

		arrayContent, err := scheYAML(rootSchema.Items, cfg)
		if err != nil {
			return nil, err
		}

		result.Content = []*yaml.Node{arrayContent}

	case NullValue:
		result.Kind = yaml.ScalarNode
		result.Value = NullValue

	// Leftover options: string, number, integer, boolean
	default:
		result.Kind = yaml.ScalarNode

		// derive a schema with default from the highest specificity (the rootschema) to lower (pattern properties in order)
		schemas := append([]*jsonschema.Schema{rootSchema}, cfg.PatternProperties...)
		if schema, ok := coalesce(schemas, withDefault); ok {
			rootSchema = schema
		}

		if cfg.HasOverride && (all(schemas, nullable) || cfg.ValueOverride != nil) {
			if cfg.ValueOverride == nil {
				cfg.ValueOverride = NullValue
			}

			result.Value = fmt.Sprint(cfg.ValueOverride)

			break
		}

		switch {
		case rootSchema.Default != nil:
			result.Value = fmt.Sprint(rootSchema.Default)

		default:
			result.LineComment = cfg.TODOComment
			result.Value = NullValue
		}
	}

	return result, nil
}

// scheYAMLObject encapsulates the logic to scheYAML a schema of type "object"
func scheYAMLObject(schema *jsonschema.Schema, cfg *Config) ([]*yaml.Node, error) { //nolint:gocyclo,cyclop,gocognit // Acceptable complexity, splitting this up is overkill
	// exit early if either schema or config is not defined
	if schema == nil || cfg == nil {
		return nil, fmt.Errorf("nil schema or config supplied: %w", ErrParsing)
	}

	// guard that all regexes are valid
	if schema.PatternProperties != nil && len(*schema.PatternProperties) > 0 {
		for pattern := range *schema.PatternProperties {
			if _, err := regexp.Compile(pattern); err != nil {
				return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
			}
		}
	}

	// properties is the join of the schema properties and the supplied overrides (which potentially match pattern properties)
	var properties []string
	if p := schema.Properties; p != nil && len(*p) > 0 {
		properties = append(properties, slices.Collect(maps.Keys(*p))...)
	}
	if overrides := cfg.ValueOverrides; len(overrides) > 0 {
		properties = append(properties, slices.Collect(maps.Keys(overrides))...)
	}
	if inherited := cfg.PatternProperties; len(inherited) > 0 {
		for _, patternschema := range inherited {
			if p := patternschema.Properties; p != nil && len(*p) > 0 {
				properties = append(properties, slices.Collect(maps.Keys(*p))...)
			}
		}
	}
	properties = unique(properties)
	sort.Strings(properties)

	// exit early if nothing matches with an empty object definition
	if len(properties) == 0 {
		return []*yaml.Node{{Kind: yaml.MappingNode, Value: "{}"}}, nil
	}

	result := make([]*yaml.Node, 0, 2*len(properties)) //nolint:mnd // not a magic number, nodes come in pairs of key=node
	for _, propertyName := range properties {
		override, hasOverride := cfg.overrideFor(propertyName)
		// if running in onlyRequired mode, emit required properties and overrides only
		if _, ok := cfg.ValueOverrides[propertyName]; !ok && cfg.OnlyRequired && !required(schema, propertyName) {
			continue
		} else if hasOverride && override == SkipValue {
			// or if an override is supplied but it is the skip sentinel, continue
			continue
		}

		// collect property, the patterns the propertyName matches and combine them as a slice of schemas
		// in order of specificity (property > patterns > inherited patterns)
		property := (*schema.Properties)[propertyName]
		patterns := patternProperties(schema, propertyName)
		schemas := append([]*jsonschema.Schema{property}, patterns...)

		// resolve potential references in schemas
		schemas = resolve(schemas)

		if inherited := cfg.PatternProperties; len(inherited) > 0 {
			for _, patternschema := range inherited {
				if patternProperty, hasProperty := (*patternschema.Properties)[propertyName]; hasProperty {
					schemas = append(schemas, patternProperty)
				}
			}
		}
		rootschema, _ := coalesce(schemas, notNil)
		if rootschema == nil {
			continue // as this property is not contained in properties OR pattern properties and thus invalid
		}

		// keyNode of the key: value pair in YAML
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: propertyName,
		}

		// add a HeadComment to the schema if a node is found which has a description or examples
		schemaWithDescription, hasDescription := coalesce(schemas, withDescription)
		schemaWithExamples, hasExamples := coalesce(schemas, withExamples)
		switch {
		case hasDescription && hasExamples:
			keyNode.HeadComment = formatHeadComment(*schemaWithDescription.Description, schemaWithExamples.Examples, cfg.LineLength)
		case hasDescription:
			keyNode.HeadComment = formatHeadComment(*schemaWithDescription.Description, []any{}, cfg.LineLength)
		case hasExamples:
			keyNode.HeadComment = formatHeadComment("", schemaWithExamples.Examples, cfg.LineLength)
		}

		// else recursively determine the nodeValue using scheYAML
		valueNode, err := scheYAML(rootschema, cfg.forProperty(propertyName, patterns))
		if err != nil {
			return nil, fmt.Errorf("failed to scheyaml %q: %w", propertyName, err)
		}

		// in case only
		if len(valueNode.Content) == 0 && valueNode.Kind == yaml.MappingNode {
			valueNode.Value = "{}"
		}

		result = append(result, keyNode, valueNode)
	}

	return result, nil
}

// resolve returns a new slice in which schemas that are references are replaced with the resolved reference
func resolve(schemas []*jsonschema.Schema) []*jsonschema.Schema {
	if len(schemas) == 0 {
		return schemas
	}

	res := make([]*jsonschema.Schema, len(schemas))
	for i, s := range schemas {
		if s != nil && s.Ref != "" {
			res[i] = s.ResolvedRef
			continue
		}
		res[i] = s
	}

	return res
}

// patternProperties returns matching pattern properties sorted in alphabetical order for some property name
func patternProperties(schema *jsonschema.Schema, propertyName string) []*jsonschema.Schema {
	patterns := schema.PatternProperties
	if patterns == nil || len(*patterns) == 0 {
		return []*jsonschema.Schema{}
	}

	result := make([]*jsonschema.Schema, 0, len(*patterns))
	for _, pattern := range slices.Sorted(maps.Keys(*patterns)) {
		patternschema := (*patterns)[pattern]
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(propertyName) {
			result = append(result, patternschema)
		}
	}

	return result
}

// formatHeadComment will generate the comment above the property with the description
// and example values. The description will be word-wrapped in case it exceeds the given non-zero lineLength.
func formatHeadComment(description string, examples []any, lineLength uint) string {
	var builder strings.Builder

	if description != "" {
		if lineLength > 0 {
			description = wordwrap.WrapString(description, lineLength)
		}

		// Empty new lines aren't respected by default
		description = strings.ReplaceAll(description, "\n\n", "\n#\n")

		builder.WriteString(description)
	}

	if description != "" && len(examples) > 0 {
		// Empty newlines aren't respected, so we need to add our own #
		builder.WriteString("\n#\n")
	}

	if len(examples) > 0 {
		// Have to prepend a # here, newlines aren't commented by default
		builder.WriteString("Examples:\n")
		for _, example := range examples {
			_, _ = builder.WriteString("- ")
			if example != nil {
				_, _ = builder.WriteString(fmt.Sprint(example))
			} else {
				_, _ = builder.WriteString(NullValue)
			}
			_, _ = builder.WriteRune('\n')
		}
	}

	return builder.String()
}
