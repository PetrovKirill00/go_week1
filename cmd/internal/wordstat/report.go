package wordstat

import (
	"fmt"
	"io"
)

func PrintReport(w io.Writer, entries []Entry, opts Options) error {
	for _, e := range entries {
		if _, err := fmt.Fprintln(w, e.Word, e.Count); err != nil {
			return fmt.Errorf("print report line %w", err)
		}
	}
	return nil
}
