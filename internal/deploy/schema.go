package deploy

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Schema describes a deploy target (Railway, Fly, Vercel front-end example, …).
type Schema struct {
	ID          string       `yaml:"id"`
	DisplayName string       `yaml:"displayName"`
	Description string       `yaml:"description"`
	Defaults    Defaults     `yaml:"defaults"`
	Files       []SchemaFile `yaml:"files"`
	Notes       string       `yaml:"notes"`
}

// Defaults are merged into template data; CLI flags override.
type Defaults struct {
	Port        string `yaml:"port"`
	MainPackage string `yaml:"main_package"`
	HealthPath  string `yaml:"health_path"`
	Module      string `yaml:"module"`
	APIBaseURL  string `yaml:"api_base_url"`
}

// SchemaFile maps an embedded template path to a repo-relative output path.
type SchemaFile struct {
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
}

// LoadSchema returns a single target by id.
func LoadSchema(id string) (*Schema, error) {
	schemas, err := LoadAllSchemas()
	if err != nil {
		return nil, err
	}
	for i := range schemas {
		if schemas[i].ID == id {
			return &schemas[i], nil
		}
	}
	return nil, fmt.Errorf("unknown deploy target %q (thy-case-llm deploy list)", id)
}

// LoadAllSchemas parses every YAML file under bundle/schemas.
func LoadAllSchemas() ([]Schema, error) {
	var out []Schema
	err := fs.WalkDir(bundleFS, "bundle/schemas", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
			return nil
		}
		b, err := fs.ReadFile(bundleFS, path)
		if err != nil {
			return err
		}
		var s Schema
		if err := yaml.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if s.ID == "" {
			return fmt.Errorf("%s: missing id", path)
		}
		out = append(out, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// ListTargetSummaries returns sorted id + display name lines for CLI list.
func ListTargetSummaries() ([]string, error) {
	schemas, err := LoadAllSchemas()
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(schemas))
	for _, s := range schemas {
		name := s.DisplayName
		if name == "" {
			name = s.ID
		}
		lines = append(lines, s.ID+"\t"+name)
	}
	return lines, nil
}

// Describe returns a short multi-line description for `deploy show`.
func (s *Schema) Describe() string {
	var b strings.Builder
	if s.DisplayName != "" {
		fmt.Fprintf(&b, "%s (%s)\n", s.DisplayName, s.ID)
	} else {
		fmt.Fprintf(&b, "%s\n", s.ID)
	}
	if s.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", strings.TrimSpace(s.Description))
	}
	if len(s.Files) > 0 {
		b.WriteString("\nDosyalar:\n")
		for _, f := range s.Files {
			fmt.Fprintf(&b, "  %s  ←  %s\n", f.Output, f.Template)
		}
	}
	if strings.TrimSpace(s.Notes) != "" {
		b.WriteString("\nNotlar:\n")
		b.WriteString(strings.TrimRight(s.Notes, "\n"))
		b.WriteByte('\n')
	}
	return b.String()
}
