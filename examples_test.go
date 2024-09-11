package scheyaml

import (
	"fmt"
	"log"
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

	result, err := SchemaToYAML([]byte(input))
	if err != nil {
		log.Fatal(err.Error())
		return
	}

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
	//       name: # TODO: Fill this in
	//       # The price of the product
	//       price: 4.5
	// # The name of the customer
	// name: Robin
}
