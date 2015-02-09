package main

import (
	"testing"
	"fmt"
	"bytes"
)

func TestParseBuffer(t *testing.T) {
	metricBuffer := []byte("a.b.c 12 1\na.b.c 12 1\na.b.c 12 1\na.b.c 12 1\na.b.")
	metrics, remaining := ParseBuffer(metricBuffer)

	for _, b := range metrics {
		if !bytes.Equal(b, []byte{'a', '.', 'b', '.', 'c', ' ', '1', '2', ' ', '1'}) {
			fmt.Printf("Metrics: Expected: '%x', Actual: '%x'", "a.b.c 123 1234567", b)
			t.Fail()
			return
		}
	}

	if !bytes.Equal(remaining, []byte{'a', '.', 'b', '.'}) {
		fmt.Printf("Remaining: Expected: 'a.b.', Actual: %s", string(remaining))
		t.Fail()
		return
	}

}

func BenchmarkParseBuffer(b *testing.B) {
	metricBuffer := []byte("a.b.c 12 1\na.b.c 12 1\na.b.c 12 1\na.b.c 12 1\na.b.c 12 1234\na.v 1 12\nav.b 1 234\na.vasdf.v 1 21231\nadfahsfaisfhalskdfjhasfiuahsfskldaj.asf")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseBuffer(metricBuffer)
	}
}
