package scheyaml

// PropType is a limited set of options from: https://json-schema.org/understanding-json-schema/reference/type
type PropType string

const (
	TypeString  PropType = "string"
	TypeInteger PropType = "integer"
	TypeNumber  PropType = "number"
	TypeBoolean PropType = "boolean"
	TypeNull    PropType = "null"

	TypeArray  PropType = "array"
	TypeObject PropType = "object"
)

// typeDefaultValues is a mapping of property types and string-based zero values in the output
var typeDefaultValues = map[PropType]string{
	TypeString:  "",
	TypeInteger: "0",
	TypeNumber:  "0",
	TypeBoolean: "false",

	// Unused, but to make it complete
	TypeNull:   "null",
	TypeObject: "{}",
	TypeArray:  "[]",
}

// DefaultValue returns a string representation of a zero value for the property type.
func (p PropType) DefaultValue() string {
	return typeDefaultValues[p]
}

// String turns PropType back into a string
func (p PropType) String() string {
	return string(p)
}
