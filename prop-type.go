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

// String turns PropType back into a string
func (p PropType) String() string {
	return string(p)
}
