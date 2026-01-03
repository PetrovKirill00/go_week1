package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/Absurd00x/go_week1/cmd/internal/wordstat"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	k := flag.Int("k", 0, "how many entries to print (0 = all)")
	min := flag.Int("min", 1, "minimum count to include")
	sortBy := flag.String("sort", "word", "sort by: word|count")
	flag.Parse()

	opts := wordstat.Options{
		K:      *k,
		Min:    *min,
		SortBy: *sortBy,
	}

	if opts.SortBy != "word" && opts.SortBy != "count" {
		fmt.Fprintln(os.Stderr, "invalid -sort, use word|count")
		os.Exit(2)
	}

	words, err := wordstat.ReadWords(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	counts := wordstat.Count(words)
	entries := wordstat.BuildEntries(counts)

	entries = wordstat.FilterMin(entries, opts.Min)

	wordstat.SortEntries(entries, opts)

	if opts.K > 0 && opts.K < len(entries) {
		entries = entries[:opts.K]
	}

	if err := wordstat.PrintReport(out, entries, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
