package wordstat

import (
	"bufio"
	"context"
	"sync"
)

func CountBufioConcurrent(ctx context.Context, in *bufio.Reader, workers int, batchSize int) (map[string]int, error) {
	if workers <= 1 {
		return CountBufio(ctx, in)
	}

	if batchSize <= 0 {
		batchSize = 1024
	}

	type shard struct {
		ch chan []string
		m  map[string]int
	}

	shards := make([]shard, workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	pool := sync.Pool{
		New: func() any {
			return make([]string, 0, batchSize)
		},
	}

	for i := 0; i < workers; i++ {
		shards[i] = shard{
			ch: make(chan []string, 8),
			m:  make(map[string]int),
		}

		go func(sh *shard) {
			defer wg.Done()
			for batch := range sh.ch {
				for _, w := range batch {
					sh.m[w]++
				}
				batch = batch[:0]
				pool.Put(batch)
			}
		}(&shards[i])
	}

	closeAll := func() {
		for i := 0; i < workers; i++ {
			close(shards[i].ch)
		}
		wg.Wait()
	}

	bufs := make([][]string, workers)
	for i := 0; i < workers; i++ {
		bufs[i] = pool.Get().([]string)[:0]
	}

	flush := func(i int) error {
		if len(bufs[i]) == 0 {
			return nil
		}

		batch := bufs[i]
		bufs[i] = pool.Get().([]string)[:0]

		select {
		case <-ctx.Done():
			return ctx.Err()
		case shards[i].ch <- batch:
			return nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			closeAll()
			return nil, ctx.Err()
		default:
		}

		s, ok, err := ReadWord(in)
		if err != nil {
			closeAll()
			return nil, err
		}
		if !ok {
			break
		}

		idx := int(hash32(s) % uint32(workers))
		bufs[idx] = append(bufs[idx], s)
		if len(bufs[idx]) >= batchSize {
			if err := flush(idx); err != nil {
				closeAll()
				return nil, err
			}
		}
	}

	for i := 0; i < workers; i++ {
		if err := flush(i); err != nil {
			closeAll()
			return nil, err
		}
	}

	closeAll()

	out := make(map[string]int)
	for i := 0; i < workers; i++ {
		for k, v := range shards[i].m {
			out[k] += v
		}
	}
	return out, nil
}

func hash32(s string) uint32 {
	// FNV-1a 32-bit
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
