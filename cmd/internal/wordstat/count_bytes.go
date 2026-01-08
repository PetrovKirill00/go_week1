package wordstat

import (
	"context"
	"fmt"
	"io"
	"strings"
	"unsafe"
)

func CountReaderBuffered(ctx context.Context, r io.Reader) (map[string]int, error) {
	data, err := io.ReadAll(ctxReader{ctx: ctx, r: r})
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}
	return CountBytes(ctx, data)
}

func CountBytes(ctx context.Context, data []byte) (map[string]int, error) {
	counts := make(map[string]int)

	i := 0
	for i < len(data) {
		if i&0xFFFF == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}
		for i < len(data) && isSpace(data[i]) {
			i++
		}
		if i >= len(data) {
			break
		}
		start := i
		for i < len(data) && !isSpace(data[i]) {
			i++
		}
		end := i

		if end-start >= 3 && data[start] == 0xEF && data[start+1] == 0xBB && data[start+2] == 0xBF {
			start += 3
			if start >= end {
				continue
			}
		}

		ascii := true
		for j := start; j < end; j++ {
			b := data[j]
			if b >= 0x80 {
				ascii = false
				break
			}
			if 'A' <= b && b <= 'Z' {
				data[j] = b + ('a' - 'A')
			}
		}

		if ascii {
			// zero-copy string for LOOKUP ONLY
			tok := unsafe.String(&data[start], end-start)

			if c, ok := counts[tok]; ok {
				counts[tok] = c + 1
			} else {
				// Новый ключ кладём как безопасную копию
				key := string(data[start:end])
				counts[key] = 1
			}
			continue
		}

		s := string(data[start:end])
		s = strings.TrimPrefix(s, "\ufeff")
		s = strings.ToLower(s)
		counts[s]++
	}

	return counts, nil
}
