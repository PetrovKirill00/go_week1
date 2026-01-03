package wordstat

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	if got := Normalize("HeLLo"); got != "hello" {
		t.Fatalf("Normalize() = %q, want %q", got, "hello")
	}
}

func TestReadWords(t *testing.T) {
	in := strings.NewReader("aa bb aa")
	words, err := ReadWords(in)
	if err != nil {
		t.Fatalf("ReadWords() error = %v", err)
	}
	if len(words) != 3 {
		t.Fatalf("len(words) = %d, want 3", len(words))
	}
}
