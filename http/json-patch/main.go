package main

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

func main() {
	fmt.Println("simple usecase:")
	simple()

	fmt.Println("\nexample 1:")
	example1()

	fmt.Println("\nexample 2:")
	example2()

}

func simple() {

	// Let's create a merge patch from these two documents...
	original := []byte(`{"name": "John", "age": 24, "height": 3.21, "weight": 80}`)
	patch := []byte(`{"name": "Jane", "age": 19, "weight": null}`)

	// Now lets apply the patch against a different JSON document...
	modified, err := jsonpatch.MergePatch(original, patch)
	if err != nil {
		panic(err)
	}

	fmt.Printf("original document:   %s\n", original)
	fmt.Printf("patch document:      %s\n", patch)
	fmt.Printf("modified document:   %s\n", modified)
	// RESULT >
	// original document:   {"name": "John", "age": 24, "height": 3.21}
	// patch document:      {"name": "Jane", "age": 19}
	// modified document:   {"name":"Jane","age":19,"height":3.21}

}

func example1() {

	// Let's create a merge patch from these two documents...
	original := []byte(`{"name": "John", "age": 24, "height": 3.21, "weight": 80}`)
	target := []byte(`{"name": "Jane", "age": 24, "weight": null}`)

	patch, err := jsonpatch.CreateMergePatch(original, target)
	if err != nil {
		panic(err)
	}

	fmt.Printf("patch document:   %s\n", patch)
	// RESULT >
	// patch document:   {"height":null,"name":"Jane"}

	// Now lets apply the patch against a different JSON document...
	alternative := []byte(`{"name": "Tina", "age": 28, "height": 3.75}`)
	modifiedAlternative, err := jsonpatch.MergePatch(alternative, patch)
	if err != nil {
		panic(err)
	}

	fmt.Printf("original document:   %s\n", original)
	fmt.Printf("patch document:      %s\n", patch)
	fmt.Printf("modifiedAlt doc:     %s\n", modifiedAlternative)
	// RESULT >
	// original document:   {"name": "John", "age": 24, "height": 3.21}
	// patch document:   {"height":null,"name":"Jane"}
	// updated alternative doc: {"name":"Jane","age":28}

}

func example2() {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	patchJSON := []byte(`[
		{"op": "replace", "path": "/name", "value": "Jane"},
		{"op": "remove", "path": "/height"}
	]`)

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		panic(err)
	}

	modified, err := patch.Apply(original)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Original document: %s\n", original)
	fmt.Printf("Modified document: %s\n", modified)
	// RESULT >
	// Original document: {"name": "John", "age": 24, "height": 3.21}
	// Modified document: {"name":"Jane","age":24}
}
