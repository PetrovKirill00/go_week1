package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/PetrovKirill00/go_week1/cmd/internal/wordstat"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	k := flag.Int("k", 0, "how many entries to print (0 = all)")
	min := flag.Int("min", 1, "minimum count to include")
	sortBy := flag.String("sort", "word", "sort by: word|count")
	format := flag.String("format", "text", "output format: text|json")
	workers := flag.Int("workers", 1, "number of counting workers (>=1)")
	flag.Parse()

	opts := wordstat.Options{
		K:       *k,
		Min:     *min,
		SortBy:  *sortBy,
		Format:  *format,
		Workers: *workers,
	}

	paths := flag.Args()

	var in io.Reader = os.Stdin

	if len(paths) > 0 {
		readers := make([]io.Reader, 0, len(paths)*2-1)
		closers := make([]io.Closer, 0, len(paths))

		closeAll := func() {
			for i := len(closers) - 1; i >= 0; i-- {
				_ = closers[i].Close()
			}
		}

		for i, p := range paths {
			f, err := os.Open(p)
			if err != nil {
				closeAll()
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			closers = append(closers, f)
			readers = append(readers, f)

			if i+1 < len(paths) {
				readers = append(readers, strings.NewReader("\n"))
			}
		}
		defer closeAll()

		in = io.MultiReader(readers...)
	}

	if err := wordstat.Run(in, out, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
