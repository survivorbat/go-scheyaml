package scheyaml

import (
	"slices"

	"github.com/kaptinlin/jsonschema"
	"golang.org/x/exp/maps"
)

// nullable iff the schema is not nil, has only two types where the second type is 'null'
func nullable(schema *jsonschema.Schema) bool {
	return schema != nil && len(schema.Type) == 2 && schema.Type[1] == "null"
}

// required returns true iff the propertyName is contained in the required property slice
func required(schema *jsonschema.Schema, propertyName string) bool {
	return slices.Contains(schema.Required, propertyName)
}

// withDefault returns the first schema which is not nil and which has a default value
func withDefault(schema *jsonschema.Schema) bool {
	return schema != nil && schema.Default != nil
}

// withDescription returns the first schema that is not nil and for which the description is non-empty
func withDescription(schema *jsonschema.Schema) bool {
	return schema != nil && schema.Description != nil && *schema.Description != ""
}

// withExamples returns the first schema which has examples
func withExamples(schema *jsonschema.Schema) bool {
	return schema != nil && len(schema.Examples) > 0
}

// notNil returns true if a received pointer to some element E is not nil
func notNil[E any](element *E) bool {
	return element != nil
}

// unique returns a new slice in which duplicate keys are removed (potentially a smaller slice)
func unique[S ~[]E, E comparable](s S) []E {
	if len(s) < 2 { //nolint:mnd // a slice of length 0 or 1 is implicitly unique
		return s
	}

	track := map[E]bool{}
	for _, e := range s {
		track[e] = true
	}

	return maps.Keys(track)
}

// coalesce returns the first value that matches the predicate or _, false if no value matches the predicate
func coalesce[S ~[]E, E any](s S, predicate func(E) bool) (E, bool) { //nolint:ireturn // type known by caller
	var orElse E
	if len(s) == 0 {
		return orElse, false
	}

	for _, element := range s {
		if predicate(element) {
			return element, true
		}
	}

	return orElse, false
}

// all returns true iff the slice is empty OR if any of the elements matches the predicate
func all[S ~[]E, E any](s S, predicate func(E) bool) bool {
	if len(s) == 0 {
		return true
	}

	for _, element := range s {
		if predicate(element) {
			return true
		}
	}

	return false
}
