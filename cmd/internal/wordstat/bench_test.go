package wordstat

import (
	"bufio"
	"context"
	"math/rand"
	"strings"
	"testing"
)

var sink int

func GenerateBenchInput(k, n int, minLen, maxLen int, seed int64) string {
	if k <= 0 {
		k = 1
	}
	if n < 0 {
		n = 0
	}
	if minLen <= 0 {
		minLen = 3
	}
	if maxLen < minLen {
		maxLen = minLen
	}
	rng := rand.New(rand.NewSource(seed))
	dict := make([]string, k)
	totalLen := -1
	for i := 0; i < k; i++ {
		L := minLen
		if maxLen > minLen {
			L += rng.Intn(maxLen - minLen + 1)
		}
		totalLen += L + 1
		b := make([]byte, L)
		for j := 0; j < L; j++ {
			b[j] = byte('a' + rng.Intn(26))
		}
		dict[i] = string(b)
	}
	var sb strings.Builder
	sb.Grow(totalLen + 10)

	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(dict[rng.Intn(k)])
	}
	return sb.String()
}

var k int = 1000
var n int = 1_400_000
var minLen int = 4
var maxLen int = 12
var seed int64 = 1

func BenchmarkCountSequential(b *testing.B) {
	input := GenerateBenchInput(k, n, minLen, maxLen, seed)
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		in := bufio.NewReader(strings.NewReader(input))
		m, err := CountBufio(context.Background(), in)
		if err != nil {
			b.Fatal(err)
		}
		sink = len(m)
	}
}

func BenchmarkCountConcurrent(b *testing.B) {
	input := GenerateBenchInput(k, n, minLen, maxLen, seed)
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		in := bufio.NewReader(strings.NewReader(input))
		m, err := CountBufioConcurrent(context.Background(), in, 4, 1024)
		if err != nil {
			b.Fatal(err)
		}
		sink = len(m)
	}
}

func BenchmarkCountBuffered(b *testing.B) {
	input := GenerateBenchInput(k, n, minLen, maxLen, seed)
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m, err := CountReaderBuffered(context.Background(), strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
		sink = len(m)
	}
}
