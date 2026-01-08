package wordstat

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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

func TestRun(t *testing.T) {
	in := strings.NewReader("b a a b c")
	var out strings.Builder

	opts := Options{K: 0, Min: 1, SortBy: "count"}
	if err := Run(in, &out, opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := "a 2\nb 2\nc 1\n"
	if out.String() != want {
		t.Fatalf("got:\n%q\nwant:\n%q", out.String(), want)
	}
}

func TestValidateOptions(t *testing.T) {
	if err := ValidateOptions(Options{SortBy: "wat"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRun_Table(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name:  "min filters",
			input: "a a b",
			opts:  Options{Min: 2, SortBy: "word"},
			want:  "a 2\n",
		},
		{
			name:    "k limits",
			input:   "a a b c",
			opts:    Options{Min: 1, SortBy: "wat"},
			wantErr: true,
		},
		{
			name:    "bad sort",
			input:   "a a b",
			opts:    Options{Min: 1, SortBy: "wat"},
			wantErr: true,
		},
		{
			name:  "bom is trimmed",
			input: "\ufeffa a b",
			opts:  Options{Min: 1, SortBy: "count"},
			want:  "a 2\nb 1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			var out strings.Builder

			err := Run(in, &out, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expectod error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if out.String() != tt.want {
				t.Fatalf("got:\n%q\nwant:\n%q", out.String(), tt.want)
			}
		})
	}
}

func TestRunJSON(t *testing.T) {
	in := strings.NewReader("b a a b c")
	var out strings.Builder

	opts := Options{SortBy: "count", Min: 1, Format: "json"}
	if err := Run(in, &out, opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var got []Entry
	if err := json.Unmarshal([]byte(out.String()), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, out=%q", err, out.String())
	}

	want := []Entry{
		{Word: "a", Count: 2},
		{Word: "b", Count: 2},
		{Word: "c", Count: 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want %v", got, want)
	}
}

func TestRunCtx_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := strings.NewReader("a a b")
	var out strings.Builder

	err := RunCtx(ctx, in, &out, Options{SortBy: "word", Min: 1, Format: "text"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled")
	}
}
