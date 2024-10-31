package scheyaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				TODOComment:    "abc",
				LineLength:     20,
				OnlyRequired:   true,
				ValueOverrides: map[string]any{},
			},
		},
		"non-existing property returns empty ValueOverrides": {
			input:        &Config{},
			propertyName: "does-not-exist",

			expected: &Config{
				ValueOverrides: map[string]any{},
			},
		},
		"property that is not a map[string]any returns empty ValueOverrides": {
			input: &Config{
				ValueOverrides: map[string]any{"wrong-type": "abc"},
			},
			propertyName: "wrong-type",

			expected: &Config{
				ValueOverrides: map[string]any{},
			},
		},
		"subproperty is returned as expected": {
			input: &Config{
				ValueOverrides: map[string]any{"foo": map[string]any{"bar": "baz"}},
			},
			propertyName: "foo",

			expected: &Config{
				ValueOverrides: map[string]any{"bar": "baz"},
			},
		},
		"subproperty is returned with OnlyRequired=true if set on parent": {
			input: &Config{
				OnlyRequired: true,
			},
			propertyName: "foo",

			expected: &Config{
				OnlyRequired:   true,
				ValueOverrides: map[string]any{},
			},
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			result := testData.input.forProperty(testData.propertyName)

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

func TestConfig_OverrideFor_ReturnsFalseOnNestedValue(t *testing.T) {
	t.Parallel()
	// Arrange
	cfg := NewConfig()
	cfg.ValueOverrides = map[string]any{
		"abc": map[string]any{},
	}

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
