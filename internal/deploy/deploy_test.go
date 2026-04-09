package deploy

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAllSchemas(t *testing.T) {
	schemas, err := LoadAllSchemas()
	if err != nil {
		t.Fatal(err)
	}
	if len(schemas) < 4 {
		t.Fatalf("expected at least 4 schemas, got %d", len(schemas))
	}
	ids := make(map[string]bool)
	for _, s := range schemas {
		ids[s.ID] = true
	}
	for _, want := range []string{"railway", "fly", "vercel", "docker"} {
		if !ids[want] {
			t.Errorf("missing schema id %q", want)
		}
	}
}

func TestInitRailwayDryRun(t *testing.T) {
	var buf bytes.Buffer
	err := Init("railway", InitOptions{DryRun: true, OutDir: t.TempDir(), OutputWriter: &buf})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Dockerfile") {
		t.Fatalf("dry-run output should mention Dockerfile, got:\n%s", out)
	}
	if !strings.Contains(out, "FROM golang") {
		t.Fatalf("expected Dockerfile content in dry-run")
	}
}

func TestInitDockerDryRun(t *testing.T) {
	var buf bytes.Buffer
	err := Init("docker", InitOptions{DryRun: true, OutDir: t.TempDir(), OutputWriter: &buf})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "docker-compose.yml") {
		t.Fatalf("dry-run output should mention docker-compose.yml, got:\n%s", out)
	}
	if !strings.Contains(out, "services:") {
		t.Fatalf("expected compose content in dry-run")
	}
}

func TestInitConflict(t *testing.T) {
	dir := t.TempDir()
	opts := InitOptions{OutDir: dir, Module: "example.com/test"}
	if err := Init("railway", opts); err != nil {
		t.Fatal(err)
	}
	err := Init("railway", opts)
	if err == nil {
		t.Fatal("expected error when files exist without --force")
	}
	if !strings.Contains(err.Error(), "zaten var") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	opts := InitOptions{OutDir: dir, Module: "example.com/test"}
	if err := Init("railway", opts); err != nil {
		t.Fatal(err)
	}
	opts.Force = true
	if err := Init("railway", opts); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("FROM golang")) {
		t.Fatalf("unexpected Dockerfile after force: %s", b)
	}
}

func TestDetectModule(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mod, err := DetectModule(dir)
	if err != nil {
		t.Fatal(err)
	}
	if mod != "github.com/foo/bar" {
		t.Fatalf("got %q", mod)
	}
}
