package scheyaml

import (
	"reflect"

	"github.com/kaptinlin/jsonschema"
)

// Option is used to customise the output, we currently don't allow extensions yet
type Option func(*Config)

// defaultTODOComment is used if no default value was defined or a property
const defaultTODOComment = "TODO: Fill this in"

// defaultLineLength is a reasonable 80 characters to make sure the generated
// output isn't overly long, and schema files don't need to insert their own
// newlines.
//
// Also, this docstring is written to not exceed the 80 character limit :)
const defaultLineLength = 80

// NewConfig instantiates a config object with default values
func NewConfig() *Config {
	return &Config{
		ValueOverrides: make(map[string]any),
		TODOComment:    defaultTODOComment,
		LineLength:     defaultLineLength,
	}
}

// Config serves as the configuration object to allow customisation in the library
type Config struct {
	// ValueOverride is a primitive value used outside of objects (mainly to support ItemsOverrides). For
	// example when the schema is an array of which the items are primitives (e.g. "string"), the overrides
	// are processed per item (hence the relation to ItemsOverrides) but are not of type map[string]any
	// that is commonly used in ValueOverrides.
	ValueOverride any

	// HasOverride is configured in conjunction with ValueOverride to distinguish between the explicit
	// and implicit nil
	HasOverride bool

	// ValueOverrides allows a user to override the default values of a schema with the given value(s).
	// Because a schema may nested, this takes the form of a map[string]any of which the structure must mimic
	// the schema to function.
	ValueOverrides map[string]any

	// ItemsOverrides allows a user to override the default values of a schema with the given value(s).
	// Because a schema may be a slice (of potentially nested maps) this is stored separately from ValueOverrides
	ItemsOverrides []any

	// PatternProperties inherited from parent
	PatternProperties []*jsonschema.Schema

	// TODOComment is used in case no default value was defined for a property. It is set by
	// default in NewConfig but can be emptied to remove the comment altogether.
	TODOComment string

	// OnlyRequired properties are returned
	OnlyRequired bool

	// LineLength prevents descriptions and unreasonably long lines. Can be disabled
	// completely by setting it to 0.
	LineLength uint

	// Indent used when marshalling to YAML
	Indent int

	// SkipValidate of the provided jsonschema and override values. Might result in undefined behavior, use
	// at own risk.
	SkipValidate bool
}

// forProperty will construct a config object for the given property, allows for recursive
// digging into property overrides
func (c *Config) forProperty(propertyName string, patternProps []*jsonschema.Schema) *Config { //nolint:cyclop // accepted complexity for forProperty
	var valueOverride any
	var hasValueOverride bool
	var valueOverrides map[string]any
	var itemsOverrides []any

	propertyOverrides, ok := c.ValueOverrides[propertyName]
	if mapoverrides, isMapStringAny := asMapStringAny(propertyOverrides); ok && isMapStringAny {
		valueOverrides = mapoverrides
	} else if sliceoverrides, isSliceMapStringAny := asSliceAny(propertyOverrides); ok && isSliceMapStringAny {
		itemsOverrides = sliceoverrides
	} else if ok {
		valueOverride = propertyOverrides
		hasValueOverride = true
	}

	patterns := make([]*jsonschema.Schema, 0, len(patternProps)+len(c.PatternProperties))
	patterns = append(patterns, patternProps...)
	for _, p := range c.PatternProperties {
		if len(p.Type) > 0 && p.Type[0] == "object" {
			// add property
			if p.Properties != nil && len(*p.Properties) > 0 {
				if property, hasProperty := (*p.Properties)[propertyName]; hasProperty {
					patterns = append(patterns, property)
				}
			}

			patterns = append(patterns, patternProperties(p, propertyName)...)
		}
	}

	if valueOverrides == nil {
		valueOverrides = make(map[string]any)
	}

	if itemsOverrides == nil {
		itemsOverrides = make([]any, 0)
	}

	return &Config{
		ValueOverride:     valueOverride,
		HasOverride:       hasValueOverride,
		ValueOverrides:    valueOverrides,
		ItemsOverrides:    itemsOverrides,
		PatternProperties: patterns,
		TODOComment:       c.TODOComment,
		OnlyRequired:      c.OnlyRequired,
		LineLength:        c.LineLength,
	}
}

// forIndex will construct a config object for the given index, allows for recursive
// digging into property overrides for
func (c *Config) forIndex(index int) *Config {
	var valueOverride any
	var hasValueOverride bool
	var valueOverrides map[string]any

	if len(c.ItemsOverrides) > index {
		if value, asMap := asMapStringAny(c.ItemsOverrides[index]); asMap {
			valueOverrides = value
		} else {
			valueOverride = c.ItemsOverrides[index]
			hasValueOverride = true
		}
	}

	return &Config{
		HasOverride:       hasValueOverride,
		ValueOverride:     valueOverride,
		ValueOverrides:    valueOverrides,
		ItemsOverrides:    nil,
		PatternProperties: nil,
		TODOComment:       c.TODOComment,
		OnlyRequired:      c.OnlyRequired,
		LineLength:        c.LineLength,
	}
}

// overrideFor examines ValueOverrides to see if there are any override values defined for the given
// propertyName. It will not return nested map[string]any values (or aliasses thereof) nor []any or
// aliases thereof. These are resolved in overrideForIndex
func (c *Config) overrideFor(propertyName string) (any, bool) {
	// Does it exist
	propertyOverride, ok := c.ValueOverrides[propertyName]
	if !ok {
		return nil, false
	}

	// Is it ~map[string]any
	if _, isMapStringAny := asMapStringAny(propertyOverride); isMapStringAny {
		return nil, false
	}

	// Is it ~[]any
	if _, isSliceAny := asSliceAny(propertyOverride); isSliceAny {
		return nil, false
	}

	if _, shouldSkip := propertyOverride.(skipValue); shouldSkip {
		return SkipValue, true
	}

	return propertyOverride, true
}

// asSliceAny returns the input converted to []any, true if the input can be represented
// as a []any (either directly or with reflect) or nil, false otherwise
func asSliceAny(input any) ([]any, bool) {
	if input == nil {
		return nil, false
	}

	// try without reflect
	if value, isSlice := input.([]any); isSlice {
		return value, true
	} else if reflect.TypeOf(input).ConvertibleTo(reflect.TypeOf([]any{})) {
		return reflect.ValueOf(input).Convert(reflect.TypeOf([]any{})).Interface().([]any), true //nolint:forcetypeassert // converted type
	}

	// fallback to reflect
	if reflect.TypeOf(input).Kind() == reflect.Slice {
		valueOf := reflect.ValueOf(input)
		res := make([]any, 0, valueOf.Len())
		for _, value := range valueOf.Seq2() {
			res = append(res, value.Interface())
		}

		return res, true
	}

	return nil, false
}

// asMapStringAny returns the input represented as map[string]any, true if the input
// can be converted (either directly or with reflect) or nil, false otherwise.
func asMapStringAny(input any) (map[string]any, bool) {
	if input == nil {
		return nil, false
	}

	if value, isMap := input.(map[string]any); isMap {
		return value, true
	} else if reflect.TypeOf(input).ConvertibleTo(reflect.TypeOf(map[string]any{})) {
		return reflect.ValueOf(input).Convert(reflect.TypeOf(map[string]any{})).Interface().(map[string]any), true //nolint:forcetypeassert // converted type
	}

	return nil, false
}

// WithOverrideValues allows you to override the default values from the JSON schema, you can
// nest map[string]any values to reach nested objects in the JSON schema.
func WithOverrideValues(values map[string]any) Option {
	return func(c *Config) {
		c.ValueOverrides = values
	}
}

// WithTODOComment allows you to set the 'TODO: Fill this in' comment in the output YAML. Can be
// set to an empty string to turn it off altogether
func WithTODOComment(comment string) Option {
	return func(c *Config) {
		c.TODOComment = comment
	}
}

// OnlyRequired properties are returned
func OnlyRequired() Option {
	return func(c *Config) {
		c.OnlyRequired = true
	}
}

// WithIndent amount of spaces to use when marshalling
func WithIndent(indent int) Option {
	return func(c *Config) {
		c.Indent = indent
	}
}

// WithCommentMaxLength prevents descriptions generating unreasonably long lines. Can be disabled
// completely by setting it to 0.
func WithCommentMaxLength(lineLength uint) Option {
	return func(c *Config) {
		c.LineLength = lineLength
	}
}

// SkipValidate will not evaluate jsonschema.Validate, might result in undefined behavior. Use at own risk
func SkipValidate() Option {
	return func(c *Config) {
		c.SkipValidate = true
	}
}
