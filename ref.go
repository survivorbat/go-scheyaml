package scheyaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotSupported     = errors.New("this feature is currently not supported")
	ErrInvalidReference = errors.New("failed to find or parse a reference in the document")
)

// lookup is used internal to resolve refs such as #/definitions/MySchema in the 'definitions' field. It
// is currently not capable of fetching foreign schemas from URLs.
func lookupSchemaRef(schema *JSONSchema, lookupString string) (*JSONSchema, error) {
	if !strings.HasPrefix(lookupString, "#") {
		return nil, fmt.Errorf("not looking up $ref '%s': %w", lookupString, ErrNotSupported)
	}

	// Using `misc` for this, as we can't anticipate where the definitons might be hiding
	definedSchema, err := lookup(schema.misc, strings.Split(lookupString, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed to lookup '%s': %w", lookupString, err)
	}

	var resultSchema *JSONSchema

	//nolint:errchkjson // It was already in JSONSchema, so it shouldn't be possible for this to throw errors
	rawSchema, _ := json.Marshal(definedSchema)

	// This on the other hand can be completely unusable data
	if err := json.Unmarshal(rawSchema, &resultSchema); err != nil {
		return nil, fmt.Errorf("failed to parse schema in '%s': %w", lookupString, errors.Join(ErrInvalidReference, err))
	}

	return resultSchema, nil
}

// lookup walks through the pathSegments in pursuit of the value we're seeking, there might be a library
// for this that does it more efficiently, but I wasn't able to find one.
//
//nolint:cyclop // Don't find it worth breaking this down
func lookup(object map[string]any, pathSegments []string) (any, error) {
	switch {
	// Reached the final segment, return the object found
	case len(pathSegments) == 0:
		return object, nil

	// If a lone # is input, we return the object itself
	case len(pathSegments) == 1 && pathSegments[0] == "#":
		return object, nil

	// If this is the first segment and it has a #, we recursively traverse the next segment
	case len(pathSegments) > 1 && pathSegments[0] == "#":
		return lookup(object, pathSegments[1:])

	// Reached the final segment
	case len(pathSegments) == 1:
		currentSegment := pathSegments[0]

		if value, ok := object[currentSegment]; ok {
			return value, nil
		}

		return nil, fmt.Errorf("failed to find '%s': %w", currentSegment, ErrInvalidReference)

	default:
		currentSegment := pathSegments[0]

		value, ok := object[currentSegment]
		if !ok {
			return nil, fmt.Errorf("failed to find '%s': %w", currentSegment, ErrInvalidReference)
		}

		if valueMap, ok := value.(map[string]any); ok {
			return lookup(valueMap, pathSegments[1:])
		}

		return nil, fmt.Errorf("segment '%s' was not an object: %w", currentSegment, ErrInvalidReference)
	}
}
