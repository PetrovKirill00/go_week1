package wordstat

import (
	"bufio"
	"context"
)

func CountBufio(ctx context.Context, in *bufio.Reader) (map[string]int, error) {
	counts := make(map[string]int)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		s, ok, err := ReadWord(in)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		counts[s]++
	}
	return counts, nil
}
