package scheyaml

import (
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ForProperty_ReturnsExpectedConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input        *Config
		propertyName string

		expected *Config
	}{
		"copies over 'simple' values from parent config": {
			input: &Config{
				TODOComment:  "abc",
				LineLength:   20,
				OnlyRequired: true,
			},
			propertyName: "foo",

			expected: &Config{
				HasOverride:       false,
				ValueOverride:     nil,
				ValueOverrides:    map[string]any{},
				ItemsOverrides:    []any{},
				PatternProperties: []*jsonschema.Schema{},
				TODOComment:       "abc",
				OnlyRequired:      true,
				LineLength:        20,
			},
		},
		"non-existing property returns empty ValueOverrides": {
			input:        &Config{},
			propertyName: "does-not-exist",

			expected: &Config{
				HasOverride:       false,
				ValueOverride:     nil,
				ValueOverrides:    map[string]any{},
				ItemsOverrides:    []any{},
				PatternProperties: []*jsonschema.Schema{},
				TODOComment:       "",
				OnlyRequired:      false,
				LineLength:        0,
			},
		},
		"property that is not a map[string]any returns empty ValueOverrides": {
			input: &Config{
				ValueOverrides: map[string]any{"wrong-type": "abc"},
			},
			propertyName: "wrong-type",

			expected: &Config{
				HasOverride:       true,
				ValueOverride:     "abc",
				ValueOverrides:    map[string]any{},
				ItemsOverrides:    []any{},
				PatternProperties: []*jsonschema.Schema{},
				TODOComment:       "",
				OnlyRequired:      false,
				LineLength:        0,
			},
		},
		"subproperty is returned as expected": {
			input: &Config{
				ValueOverrides: map[string]any{"foo": map[string]any{"bar": "baz"}},
			},
			propertyName: "foo",

			expected: &Config{
				HasOverride:       false,
				ValueOverride:     nil,
				ValueOverrides:    map[string]any{"bar": "baz"},
				ItemsOverrides:    []any{},
				PatternProperties: []*jsonschema.Schema{},
				TODOComment:       "",
				OnlyRequired:      false,
				LineLength:        0,
			},
		},
		"subproperty is returned with OnlyRequired=true if set on parent": {
			input: &Config{
				OnlyRequired: true,
			},
			propertyName: "foo",

			expected: &Config{
				HasOverride:       false,
				ValueOverride:     nil,
				ValueOverrides:    map[string]any{},
				ItemsOverrides:    []any{},
				PatternProperties: []*jsonschema.Schema{},
				TODOComment:       "",
				OnlyRequired:      true,
				LineLength:        0,
			},
		},
		"items overrides returned if input is a slice": {
			input: &Config{
				OnlyRequired: true,
				ValueOverrides: map[string]any{
					"beverages": []string{"coffee", "tea"},
				},
			},
			propertyName: "beverages",

			expected: &Config{
				HasOverride:       false,
				ValueOverride:     nil,
				ValueOverrides:    map[string]any{},
				ItemsOverrides:    []any{"coffee", "tea"},
				PatternProperties: []*jsonschema.Schema{},
				TODOComment:       "",
				OnlyRequired:      true,
				LineLength:        0,
			},
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			result := testData.input.forProperty(testData.propertyName, nil)

			// Assert
			assert.Equal(t, testData.expected, result)
		})
	}
}

func TestConfig_OverrideFor_ReturnsFalseOnNotExists(t *testing.T) {
	t.Parallel()
	// Arrange
	cfg := NewConfig()

	// Act
	value, ok := cfg.overrideFor("abc")

	// Assert
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestConfig_OverrideFor_ReturnsTrueOnOverrideFound(t *testing.T) {
	t.Parallel()
	// Arrange
	cfg := NewConfig()
	cfg.ValueOverrides = map[string]any{
		"abc": "def",
	}

	// Act
	value, ok := cfg.overrideFor("abc")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "def", value)
}

func TestConfig_forProperty_MapStringAnyTypeAlias(t *testing.T) {
	t.Parallel()
	// Arrange
	type MapAlias map[string]any
	cfg := NewConfig()
	cfg.ValueOverrides = map[string]any{
		"abc": MapAlias{
			"foo": "bar",
		},
	}

	// Act
	result := cfg.forProperty("abc", nil)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, map[string]any{
		"foo": "bar",
	}, result.ValueOverrides)
}

func TestConfig_forProperty_NilMapStringAnyTypeAlias(t *testing.T) {
	t.Parallel()
	// Arrange
	type MapAlias map[string]any
	cfg := NewConfig()
	cfg.ValueOverrides = map[string]any{
		"abc": MapAlias(nil),
	}

	// Act
	result := cfg.forProperty("abc", nil)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, map[string]any{}, result.ValueOverrides)
}

func TestConfig_forProperty_NilOverride(t *testing.T) {
	t.Parallel()
	// Arrange
	cfg := NewConfig()
	cfg.ValueOverrides = map[string]any{
		"abc": nil,
	}

	// Act
	result := cfg.forProperty("abc", nil)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, map[string]any{}, result.ValueOverrides)
}
