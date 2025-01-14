# 📅 Schema YAML

Ever wanted to process a user provided YAML file constrained by a [JSON schema](https://json-schema.org)? This library will help with
validating the user input to be compliant with the JSON schema as well as render a fully qualified version (with defaults) of only
the required properties (useful when e.g. generating an initial YAML file for the user which still required attention) or all properties
to retrieve the default values as defined by the JSON schema.

It uses [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema)'s `jsonschema.Schema` struct as input.

## ⬇️ Installation

`go get github.com/survivorbat/go-scheyaml`

## Migration

In V2 schema validation is by default applied which can be a breaking change for jsonschemas that are invalid.
Previously, the schema and overrides would still be interpreted but this could lead to unpredictable behavior hence
that it is now enabled by default. To skip schema validation, run it with the `SkipValidate` option:
```
result, err := SchemaToNode(schema, SkipValidate())
```

## 📋 Usage

When override values are supplied or the json schema contains default values, the following rules apply when determining
which value to use:

1) if the schema is nullable (`"type": ["<type>", "null"]`) and an override is specified for this key, use the override
2) if the schema is not nullable and the override is not `nil`, use the override value
3) if the schema has a default (`"default": "abc"`) use the default value of the property
4) if 1..N pattern properties match, use the first pattern property which has a default value (if any)

This can be especially useful when using generated JSON/YAML structs for configuration in Go applications, e.g. 
generated from [omissis/go-jsonschema](https://github.com/omissis/go-jsonschema):
```
$ go-jsonschema --capitalization=API --extra-imports json-schema.json --struct-name-from-title -o config.go -p config
```

Given some simple schema:
```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Config",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "default": "Hello World"
    }
  }
}
```

Will generate the following (simplified) Go struct:
```go
type Config struct {
	// Name corresponds to the JSON schema field "name".
	Name string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
}
```

Given some config file that should be valid (an empty file):
```yaml
# yaml-language-server: $schema=json-schema.json

```

Normally, the default values are "lost" when unmarshalling. That's where scheyaml can output a processed
version according to the json schema of the input that can be read, in this case as if the user would
have supplied:
```yaml
# yaml-language-server: $schema=json-schema.json
name: Hello World
```

See the example tests in `./examples_test.go` for more details.

## ✅ Support

- [x] Feature to override values in output
- [x] Feature to override the comment on a missing default value
- [x] Basic types (string, number, integer, null)
- [x] Object type
- [x] Array
- [x] Refs
- [x] Pattern Properties
- [ ] AnyOf
- [ ] AllOf

## 🔭 Plans

Not much yet.
