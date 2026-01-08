package wordstat

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrintReport(w io.Writer, entries []Entry, opts Options) error {
	switch opts.Format {
	case "", "text":
		for _, e := range entries {
			if _, err := fmt.Fprintln(w, e.Word, e.Count); err != nil {
				return fmt.Errorf("print report line %w", err)
			}
		}
		return nil
	case "json":
		enc := json.NewEncoder(w)
		if err := enc.Encode(entries); err != nil {
			return fmt.Errorf("encode json report: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown format %q (use text|json)", opts.Format)
	}
}
