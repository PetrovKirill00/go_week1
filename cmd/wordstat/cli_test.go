package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildWordstat(t *testing.T) string {
	t.Helper()

	exe := "wordstat"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}

	tmp := t.TempDir()
	outPath := filepath.Join(tmp, exe)

	cmd := exec.Command("go", "build", "-o", outPath, ".")
	cmd.Dir = "."
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(b))
	}
	return outPath
}

func TestCLI_Stdin_Text(t *testing.T) {
	bin := buildWordstat(t)
	cmd := exec.Command(bin, "-sort", "count", "-format", "text")
	cmd.Stdin = strings.NewReader("b a a b c")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run error=%v stderr=%q", err, stderr.String())
	}
	want := "a 2\nb 2\nc 1\n"
	if stdout.String() != want {
		t.Fatalf("got=%q want=%q", stdout.String(), want)
	}
}

func TestCLI_BadSort_ExitCode(t *testing.T) {
	bin := buildWordstat(t)
	cmd := exec.Command(bin, "-sort", "wat")
	cmd.Stdin = strings.NewReader("a a b")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit code, gut nil stderr=%q stdout=%q", stderr.String(), stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatalf("expectod error on stderr")
	}
}
