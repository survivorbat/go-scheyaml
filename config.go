package scheyaml

// Option is used to customise the output, we currently don't allow extensions yet
type Option func(*Config)

// defaultTODOComment is used if no default value was defined or a property
const defaultTODOComment = "TODO: Fill this in"

// NewConfig instantiates a config object with default values
func NewConfig() *Config {
	return &Config{
		ValueOverrides: make(map[string]any),
		TODOComment:    defaultTODOComment,
	}
}

// Config serves as the configuration object to allow customisation in the library
type Config struct {
	// ValueOverrides allows a user to override the default values of a schema with the given value(s).
	// Because a schema may nested, this takes the form of a map[string]any of which the structure must mimic
	// the schema to function.
	ValueOverrides map[string]any

	// TODOComment is used in case no default value was defined for a property. It is set by
	// default in NewConfig but can be emptied to remove the comment altogether.
	TODOComment string

	// Minimal mode only renders properties which are required and have no default value
	Minimal bool
}

// forProperty will construct a config object for the given property, allows for recursive
// digging into property overrides
func (c *Config) forProperty(propertyName string) *Config {
	var valueOverrides map[string]any

	propertyOverrides, ok := c.ValueOverrides[propertyName]
	if ok {
		valueOverrides, _ = propertyOverrides.(map[string]any)
	}

	if valueOverrides == nil {
		valueOverrides = make(map[string]any)
	}

	return &Config{
		TODOComment:    c.TODOComment,
		Minimal:        c.Minimal,
		ValueOverrides: valueOverrides,
	}
}

// overrideFor examines ValueOverrides to see if there are any override values defined for the given
// propertyName. It will not return nested map[string]any values.
func (c *Config) overrideFor(propertyName string) (any, bool) {
	// Does it exist
	propertyOverride, ok := c.ValueOverrides[propertyName]
	if !ok {
		return nil, false
	}

	// Is it NOT map[string]any
	if _, ok = propertyOverride.(map[string]any); ok {
		return nil, false
	}

	return propertyOverride, true
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

// WithMinimal allows you to only return the minimal yaml file that needs to be filled in
// by the user (only required properties with no default value)
func WithMinimal() Option {
	return func(c *Config) {
		c.Minimal = true
	}
}
