// examples/converter_example.go
package main

import (
	"fmt"
	"log"

	"github.com/opensourceways/argus-worker/converter"
)

func main() {
	// Simple GitHub workflow YAML for testing
	yamlData := []byte(`
name: Test Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        run: echo "Checkout code"
      - name: Build
        run: echo "Building application"
  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Test
        run: echo "Running tests"
`)

	result, err := converter.ConvertWorkflow(yamlData)
	if err != nil {
		log.Fatalf("ConvertWorkflow() error = %v", err)
	}

	fmt.Println("Conversion successful!")
	fmt.Printf("Result length: %d characters\n", len(result))

	// Print first 500 characters of the result
	if len(result) > 500 {
		fmt.Println("First 500 characters of result:")
		fmt.Println(result[:500])
		fmt.Println("...")
	} else {
		fmt.Println("Result:")
		fmt.Println(result)
	}
}
