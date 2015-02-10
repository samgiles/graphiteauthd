package main

import (
	"testing"
	"fmt"
	"bytes"
)

var testKey []byte = []byte("185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969")

func TestParseBuffer(t *testing.T) {
	metricBuffer :=
	[]byte("185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.c 12 1\n185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.c 12 1\n185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.c 12 1\n185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.c 12 1\n185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.")
	metrics, remaining := ParseBuffer(metricBuffer, testKey)

	for _, b := range metrics {
		if !bytes.Equal(b, append(testKey, []byte{'.', 'b', '.', 'c', ' ', '1', '2', ' ', '1'}...)) {
			fmt.Printf("Metrics: Expected: '%x', Actual: '%x'", "185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.c 123 1234567", b)
			t.Fail()
			return
		}
	}

	if !bytes.Equal(remaining, append(testKey, []byte{'.', 'b', '.'}...)) {
		fmt.Printf("Remaining: Expected: '185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.', Actual: %s", string(remaining))
		t.Fail()
		return
	}

}

func TestParseBufferFilter(t *testing.T) {
	metricBuffer :=
	[]byte("invalid381969.b.c 12 1\n1invalid18007d1764826381969.b.c 12 1\n185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.c 12 1\n1invalid.b.c 12 1\n185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969.b.")
	metrics, _:= ParseBuffer(metricBuffer, testKey)

	if len(metrics) != 1 {
		fmt.Printf("Only one metric should've passed through the key filter, instead saw: %d\n", len(metrics))
		t.Fail()
		return
	}
}

func BenchmarkParseBuffer(b *testing.B) {
	metricBuffer := []byte("a.b.c 12 1\na.b.c 12 1\na.b.c 12 1\na.b.c 12 1\na.b.c 12 1234\na.v 1 12\nav.b 1 234\na.vasdf.v 1 21231\nadfahsfaisfhalskdfjhasfiuahsfskldaj.asf")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseBuffer(metricBuffer, testKey)
	}
}
