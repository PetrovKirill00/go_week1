package wordstat

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func isSpace(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t', '\v', '\f':
		return true
	default:
		return false
	}
}

func normalizeWordBytes(b []byte) string {
	// Trim BOM bytes
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	ascii := true
	for _, c := range b {
		if c >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		for i, c := range b {
			if 'A' <= c && c <= 'Z' {
				b[i] = c + ('a' - 'A')
			}
		}
		return string(b)
	}

	s := string(b)
	s = strings.TrimPrefix(s, "\ufeff")
	return strings.ToLower(s)
}

func ReadWord(r *bufio.Reader) (string, bool, error) {
	var c byte
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			return "", false, nil
		}
		if err != nil {
			return "", false, fmt.Errorf("read byte: %w", err)
		}
		if !isSpace(b) {
			c = b
			break
		}
	}
	var stack [64]byte
	buf := stack[:0]
	buf = append(buf, c)

	for {
		b, err := r.ReadByte()
		if err == nil {
			if isSpace(b) {
				break
			}
			buf = append(buf, b)
			continue
		}
		if err == io.EOF {
			break
		}
		return "", false, fmt.Errorf("read byte: %w", err)
	}

	return normalizeWordBytes(buf), true, nil
}
