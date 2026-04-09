//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	raw, err := os.ReadFile("docs/openapi.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read docs/openapi.yaml: %v\n", err)
		os.Exit(1)
	}

	var spec any
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "yaml parse: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("docs/openapi.json", append(out, '\n'), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write docs/openapi.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("docs/openapi.json updated")
}
