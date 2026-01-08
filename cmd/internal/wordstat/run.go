package wordstat

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

func ValidateOptions(opts Options) error {
	switch opts.SortBy {
	case "word", "count":
		// ok
	default:
		return fmt.Errorf("invalid SortBy=%q (use word|count)", opts.SortBy)
	}
	switch opts.Format {
	case "", "text", "json":
		// ok
	default:
		return fmt.Errorf("invalid -format=%q (use text|json)", opts.Format)
	}
	if opts.Workers < 1 {
		return fmt.Errorf("invalid -workers=%d (must be >= 1", opts.Workers)
	}
	return nil
}

func RunCtx(ctx context.Context, r io.Reader, w io.Writer, opts Options) error {
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	if err := ValidateOptions(opts); err != nil {
		return err
	}

	var counts map[string]int
	var err error

	if opts.Buffered {
		counts, err = CountReaderBuffered(ctx, r)
	} else {
		in := bufio.NewReader(r)

		if opts.Workers <= 1 {
			counts, err = CountBufio(ctx, in)
		} else {
			counts, err = CountBufioConcurrent(ctx, in, opts.Workers, 1024)
		}
	}

	if err != nil {
		return err
	}

	entries := BuildEntries(counts)
	entries = FilterMin(entries, opts.Min)
	SortEntries(entries, opts)

	if opts.K > 0 && opts.K < len(entries) {
		entries = entries[:opts.K]
	}

	return PrintReport(w, entries, opts)
}

func Run(r io.Reader, w io.Writer, opts Options) error {
	return RunCtx(context.Background(), r, w, opts)
}
