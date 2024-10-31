# 📅 Schema Yaml

Ever wanted to turn a JSON schema into an example YAML file? Probably not, but this library allows you
to do just that (in a limited fashion).

It uses [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema)'s `jsonschema.Schema` struct as input.

## ⬇️ Installation

`go get github.com/survivorbat/go-scheyaml`

## 📋 Usage

Check out [this example](./examples_test.go)

## ✅ Support

- [x] Feature to override values in output
- [x] Feature to override the comment on a missing default value
- [x] Basic types (string, number, integer, null)
- [x] Object type
- [x] Array
- [x] Refs
- [ ] AnyOf
- [ ] AllOf
- [ ] Pattern Properties

## 🔭 Plans

Not much yet.
