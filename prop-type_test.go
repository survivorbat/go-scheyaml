package scheyaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropType_DefaultValue_ReturnsExpectedValue(t *testing.T) {
	t.Parallel()

	for inputType, expected := range typeDefaultValues {
		t.Run(inputType.String(), func(t *testing.T) {
			t.Parallel()
			// Act
			result := inputType.DefaultValue()

			// Assert
			assert.Equal(t, expected, result)
		})
	}
}
