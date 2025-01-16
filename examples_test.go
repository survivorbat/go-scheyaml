package scheyaml

import (
	"fmt"

	"github.com/kaptinlin/jsonschema"
)

func ExampleSchemaToYAML() {
	input := `{
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "default": "Robin",
        "description": "The name of the customer"
      },
      "beverages": {
        "type": "array",
        "description": "A list of beverages the customer has consumed",
        "items": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string", 
              "description": "The name of the beverage", 
              "examples": ["Coffee", "Tea", "Cappucino"]
            },
            "price": {
              "type": "number",
              "description": "The price of the product",
              "default": 4.5
            }
          }
        }
      }
    }
  }`

	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile([]byte(input))

	result, _ := SchemaToYAML(schema)

	fmt.Println(string(result))

	// Output:
	// # A list of beverages the customer has consumed
	// beverages:
	//     - # The name of the beverage
	//       #
	//       # Examples:
	//       # - Coffee
	//       # - Tea
	//       # - Cappucino
	//       name: null # TODO: Fill this in
	//       # The price of the product
	//       price: 4.5
	// # The name of the customer
	// name: Robin
}

func ExampleSchemaToYAML_withOverrideValues() {
	input := `{
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "default": "Robin",
        "description": "The name of the customer"
      },
      "previous_orders": {
        "type": "array",
		"items": {
			"type": "string"
		},
        "description": "names of beverages the customer has consumed"
      },
      "beverages": {
        "type": "array",
        "description": "A list of beverages the customer has consumed",
        "items": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string", 
              "description": "The name of the beverage", 
              "examples": ["Coffee", "Tea", "Cappucino"]
            },
            "price": {
              "type": "number",
              "description": "The price of the product",
              "default": 4.5
            },
            "description": {
              "type": "string"
            }
          }
        }
      }
    }
  }`

	overrides := map[string]any{
		"name":            "John",
		"previous_orders": []string{"Water", "Tea"},
		"beverages": []any{
			map[string]any{"name": "Coffee"},
			map[string]any{"name": "Beer"},
		},
	}

	todoComment := "Do something with this"

	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile([]byte(input))

	result, err := SchemaToYAML(schema, WithIndent(2), WithOverrideValues(overrides), WithTODOComment(todoComment))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(result))

	// Output:
	// # A list of beverages the customer has consumed
	// beverages:
	//   - description: null # Do something with this
	//     # The name of the beverage
	//     #
	//     # Examples:
	//     # - Coffee
	//     # - Tea
	//     # - Cappucino
	//     name: Coffee
	//     # The price of the product
	//     price: 4.5
	//   - description: null # Do something with this
	//     # The name of the beverage
	//     #
	//     # Examples:
	//     # - Coffee
	//     # - Tea
	//     # - Cappucino
	//     name: Beer
	//     # The price of the product
	//     price: 4.5
	// # The name of the customer
	// name: John
	// # names of beverages the customer has consumed
	// previous_orders:
	//   - Water
	//   - Tea
}
