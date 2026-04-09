package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	out := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		out <- buf.String()
	}()

	fn()
	_ = w.Close()
	os.Stdout = old
	s := <-out
	_ = r.Close()
	return s
}

func TestPrintUsage_mentionsDockerDeployTarget(t *testing.T) {
	s := captureStdout(t, printUsage)
	if !strings.Contains(s, "deploy list") || !strings.Contains(s, "docker") {
		t.Fatalf("main usage should list docker as deploy target:\n%s", s)
	}
}

func TestPrintDeployUsage_mentionsDockerId(t *testing.T) {
	s := captureStdout(t, printDeployUsage)
	if !strings.Contains(s, "railway | fly | docker | vercel") {
		t.Fatalf("deploy usage should list docker id:\n%s", s)
	}
}

func TestPrintProviderUsage_nonEmpty(t *testing.T) {
	s := captureStdout(t, printProviderUsage)
	if !strings.Contains(s, "thy-case-llm provider") {
		t.Fatalf("provider usage header missing:\n%s", s)
	}
}

func TestPrintTemplatesUsage_nonEmpty(t *testing.T) {
	s := captureStdout(t, printTemplatesUsage)
	if !strings.Contains(s, "thy-case-llm templates") {
		t.Fatalf("templates usage header missing:\n%s", s)
	}
}
