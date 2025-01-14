package scheyaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesce_EmptySlice(t *testing.T) {
	t.Parallel()
	// Arrange
	elems := []string{}

	// Act
	elem, hasElem := coalesce(elems, func(s string) bool {
		return s == "test"
	})

	// Assert
	assert.False(t, hasElem)
	assert.Equal(t, "", elem)
}

func TestAll_EmptySlice(t *testing.T) {
	t.Parallel()
	// Arrange
	elems := []string{}

	// Act
	matches := all(elems, func(s string) bool {
		return s == "test"
	})

	// Assert
	assert.True(t, matches)
}

func TestAll_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	elems := []string{"test", "test"}

	// Act
	matches := all(elems, func(s string) bool {
		return s == "test"
	})

	// Assert
	assert.True(t, matches)
}
