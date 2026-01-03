package wordstat

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Entry struct {
	Word  string
	Count int
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
	return strings.ToLower(s)
}

func ReadWord(r *bufio.Reader) (string, bool, error) {
	var s string
	_, err := fmt.Fscan(r, &s)
	if err == io.EOF {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("scan word: %w", err)
	}
	return s, true, nil
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
		word = Normalize(word)
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
