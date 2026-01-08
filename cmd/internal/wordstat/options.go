package wordstat

type Options struct {
	K        int
	Min      int
	SortBy   string
	Format   string // "text" | "json"
	Workers  int
	Buffered bool
}
