package deploy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// InitOptions controls deploy init behavior.
type InitOptions struct {
	OutDir        string
	DryRun        bool
	Force         bool
	Module        string
	Port          string
	MainPackage   string
	HealthPath    string
	APIBaseURL   string
	OutputWriter io.Writer
}

// Init renders schema templates into OutDir (or prints when DryRun).
func Init(id string, opts InitOptions) error {
	if opts.OutputWriter == nil {
		opts.OutputWriter = os.Stdout
	}
	outDir := opts.OutDir
	if outDir == "" {
		outDir = "."
	}
	outDir = filepath.Clean(outDir)

	s, err := LoadSchema(id)
	if err != nil {
		return err
	}

	data, err := mergeTemplateData(s, opts)
	if err != nil {
		return err
	}

	for _, sf := range s.Files {
		if err := renderOne(sf, data, outDir, opts); err != nil {
			return err
		}
	}
	return nil
}

func renderOne(sf SchemaFile, data TemplateData, outDir string, opts InitOptions) error {
	tplPath := filepath.ToSlash(filepath.Join("bundle/templates", sf.Template))
	b, err := fs.ReadFile(bundleFS, tplPath)
	if err != nil {
		return fmt.Errorf("şablon okunamadı %s: %w", sf.Template, err)
	}

	t, err := template.New(filepath.Base(sf.Template)).Parse(string(b))
	if err != nil {
		return fmt.Errorf("şablon parse %s: %w", sf.Template, err)
	}

	var rendered strings.Builder
	if err := t.Execute(&rendered, data); err != nil {
		return fmt.Errorf("şablon çalıştırma %s: %w", sf.Template, err)
	}

	relOut := filepath.FromSlash(sf.Output)
	dest := filepath.Join(outDir, relOut)

	if opts.DryRun {
		fmt.Fprintf(opts.OutputWriter, "=== %s ===\n", relOut)
		opts.OutputWriter.Write([]byte(rendered.String()))
		if !strings.HasSuffix(rendered.String(), "\n") {
			opts.OutputWriter.Write([]byte("\n"))
		}
		fmt.Fprintln(opts.OutputWriter)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(dest); err == nil && !opts.Force {
		return fmt.Errorf("dosya zaten var: %s (üzerine yazmak için --force)", dest)
	}
	mode := os.FileMode(0o644)
	if err := os.WriteFile(dest, []byte(rendered.String()), mode); err != nil {
		return err
	}
	return nil
}
