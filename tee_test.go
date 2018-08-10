package xmlstream

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

func TestTeeReader(t *testing.T) {
	const stream = `<one>two</one>`
	d := xml.NewDecoder(strings.NewReader(stream))
	wb1 := &bytes.Buffer{}
	e1 := xml.NewEncoder(wb1)
	wb2 := &bytes.Buffer{}
	e2 := xml.NewEncoder(wb2)
	n, err := Copy(e2, TeeReader(d, e1))
	if err != nil {
		t.Fatalf("Unexpected error copying tokens: %q", err)
	}
	if n != 3 {
		t.Fatalf("Wrong number of tokens: want=3, got=%d", n)
	}
	err = e1.Flush()
	if err != nil {
		t.Fatalf("Unexpected error flushing e1 tokens: %q", err)
	}
	err = e2.Flush()
	if err != nil {
		t.Fatalf("Unexpected error flushing e2 tokens: %q", err)
	}

	if stream != wb1.String() {
		t.Errorf("Unexpected value in output buffer: want=%q, got=%q", stream, wb1.String())
	}
	if stream != wb2.String() {
		t.Errorf("Unexpected value in copy buffer: want=%q, got=%q", stream, wb2.String())
	}
}
