package wordstat

import (
	"bufio"
	"io"
	"strings"
)

type Entry struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func FilterMin(entries []Entry, min int) []Entry {
	if min <= 1 {
		return entries
	}
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.Count >= min {
			out = append(out, e)
		}
	}
	return out
}

func FilterMinInPlace(entries []Entry, min int) []Entry {
	dst := entries[:0]
	for _, e := range entries {
		if e.Count >= min {
			dst = append(dst, e)
		}
	}
	return dst
}

func Normalize(s string) string {
	s = strings.TrimPrefix(s, "\ufeff")
	return strings.ToLower(s)
}

func ReadWords(r io.Reader) ([]string, error) {
	in := bufio.NewReader(r)
	words := make([]string, 0)

	for {
		var s string
		s, ok, err := ReadWord(in)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		words = append(words, s)
	}
	return words, nil
}

func Count(words []string) map[string]int {
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}
	return counts
}

func BuildEntries(counts map[string]int) []Entry {
	entries := make([]Entry, 0, len(counts))
	for w, c := range counts {
		entries = append(entries, Entry{w, c})
	}
	return entries
}
