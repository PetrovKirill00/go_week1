package wordstat

import "sort"

func SortEntries(entries []Entry, opts Options) {
	switch opts.SortBy {
	case "count":
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Count != entries[j].Count {
				return entries[i].Count > entries[j].Count
			}
			return entries[i].Word < entries[j].Word
		})
	case "word":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Word < entries[j].Word
		})
	}
}
