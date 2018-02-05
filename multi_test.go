// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.
//
// Code in this file copied from the Go io package:
//
// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.GO file.

package xmlstream_test

import (
	"encoding/xml"
	"errors"
	"io"
	"runtime"
	"testing"
	"time"

	"mellium.im/xmlstream"
)

// callDepth returns the logical call depth for the given PCs.
func callDepth(callers []uintptr) (depth int) {
	frames := runtime.CallersFrames(callers)
	more := true
	for more {
		_, more = frames.Next()
		depth++
	}
	return
}

// Test that MultiReader properly flattens chained multiReaders when Read is
// called
func TestMultiReaderFlatten(t *testing.T) {
	pc := make([]uintptr, 1000) // 1000 should fit the full stack
	n := runtime.Callers(0, pc)
	var myDepth = callDepth(pc[:n])
	var readDepth int // will contain the depth from which fakeReader.Read was called
	var r xml.TokenReader = xmlstream.MultiReader(xmlstream.ReaderFunc(func() (xml.Token, error) {
		n := runtime.Callers(1, pc)
		readDepth = callDepth(pc[:n])
		return nil, errors.New("irrelevant")
	}))

	// chain a bunch of multiReaders
	for i := 0; i < 100; i++ {
		r = xmlstream.MultiReader(r)
	}

	// don't care about errors, just want to check the call-depth for Read
	r.Token()

	if readDepth != myDepth+2 { // 2 should be multiReader.Read and fakeReader.Read
		t.Errorf("multiReader did not flatten chained multiReaders: expected readDepth %d, got %d",
			myDepth+2, readDepth)
	}
}

// tokenAndEOFReader is a TokenReader that always returns the underlying token
// and an io.EOF.
type tokenAndEOFReader xml.CharData

func (t tokenAndEOFReader) Token() (xml.Token, error) {
	return xml.CharData(t), io.EOF
}

// tokenThenEOFReader returns a TokenReader that first returns (t, nil) and then
// always returns (nil, io.EOF)
func tokenThenEOFReader(t xml.Token) xmlstream.ReaderFunc {
	var read bool
	return func() (xml.Token, error) {
		if read {
			return nil, io.EOF
		}
		read = true
		return t, nil
	}
}

// Test that a reader returning (t, io.EOF) at the end of an MultiReader
// chain continues to return io.EOF on its final read, rather than
// yielding a (nil, io.EOF).
func TestMultiReaderFinalEOF(t *testing.T) {
	r := xmlstream.MultiReader(
		tokenThenEOFReader(xml.CharData("a")),
		tokenAndEOFReader(xml.CharData("b")),
	)
	tok, err := r.Token()
	if string(tok.(xml.CharData)) != "a" || err != nil {
		t.Errorf(`got %v, %v; want xml.CharData("a"), nil`, tok, err)
	}
	tok, err = r.Token()
	if string(tok.(xml.CharData)) != "b" || err != io.EOF {
		t.Errorf(`got %v, %v; want xml.CharData("b"), io.EOF`, tok, err)
	}
	tok, err = r.Token()
	if tok != nil || err != io.EOF {
		t.Errorf(`got %v, %v; want nil, io.EOF`, tok, err)
	}
}

func TestMultiReaderFreesExhaustedReaders(t *testing.T) {
	var mr xml.TokenReader
	closed := make(chan struct{})
	// The closure ensures that we don't have a live reference to r1 on our stack
	// after MultiReader is inlined (https://golang.org/issue/18819).
	// This is a work around for a limitation in liveness analysis.
	func() {
		r1 := tokenAndEOFReader("foo")
		r2 := tokenAndEOFReader("bar")
		mr = xmlstream.MultiReader(r1, r2)
		runtime.SetFinalizer(&r1, func(*tokenAndEOFReader) {
			close(closed)
		})
	}()

	tok, err := mr.Token()
	if string(tok.(xml.CharData)) != "foo" || err != nil {
		t.Errorf(`got %v, %v; want "foo", nil`, tok, err)
	}

	runtime.GC()
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for collection of r1")
	}

	tok, err = mr.Token()
	if string(tok.(xml.CharData)) != "bar" || err != io.EOF {
		t.Errorf(`got %v, %v; want "bar", io.EOF`, tok, err)
	}
}

func TestInterleavedMultiReader(t *testing.T) {
	r1 := tokenAndEOFReader("123")
	r2 := tokenAndEOFReader("456")

	mr1 := xmlstream.MultiReader(r1, r2)
	mr2 := xmlstream.MultiReader(mr1)

	// Have mr2 use mr1's []Readers.
	// Consume r1 (and clear it for GC to handle).
	tok, err := mr2.Token()
	if got := string(tok.(xml.CharData)); got != "123" || err != nil {
		t.Errorf(`mr2.Token() = (%q, %v), want ("123", nil)`, got, err)
	}

	// Consume r2 via mr1.
	// This should not panic even though mr2 cleared r1.
	tok, err = mr1.Token()
	if got := string(tok.(xml.CharData)); got != "456" || err != io.EOF {
		t.Errorf(`mr1.Token() = (%q, %v), want ("456", io.EOF)`, got, err)
	}

	tok, err = mr1.Token()
	if tok != nil || err != io.EOF {
		t.Errorf(`Second mr1.Token() = (%q, %v), want (nil, io.EOF)`, tok, err)
	}
	tok, err = mr2.Token()
	if tok != nil || err != io.EOF {
		t.Errorf(`Second mr2.Token() = (%q, %v), want (nil, io.EOF)`, tok, err)
	}
}

// bufWriter is a TokenWriter that buffers the last token written to it and
// counts how many writes have occured.
type bufWriter struct {
	b  xml.Token
	wc int
}

func (w *bufWriter) EncodeToken(t xml.Token) error {
	w.b = t
	w.wc++
	return nil
}

func (w *bufWriter) Flush() error {
	return nil
}

// errWriter is a TokenWriter that always returns the provided error on any
// write call.
type errWriter struct {
	err error
}

func (w *errWriter) EncodeToken(t xml.Token) error {
	return w.err
}

func (w *errWriter) Flush() error {
	return nil
}

func TestMultiWriter(t *testing.T) {
	t.Run("Write", func(t *testing.T) {
		b1, b2, b3 := new(bufWriter), new(bufWriter), new(bufWriter)
		mw := xmlstream.MultiWriter(b1, b2, b3)

		tok := xml.CharData("test")
		if err := mw.EncodeToken(tok); err != nil {
			t.Error("Write failed unexpectedly:", err)
		}

		if stok := string(tok); string(b1.b.(xml.CharData)) != stok || string(b2.b.(xml.CharData)) != stok || string(b3.b.(xml.CharData)) != stok {
			t.Errorf("One of the tokens is not correct. want=%v, got=%v,%v,%v", tok, b1.b, b2.b, b3.b)
		}
		if b1.wc != 1 || b2.wc != 1 || b3.wc != 1 {
			t.Errorf("Expected three single writes, got: %d, %d, %d", b1.wc, b2.wc, b3.wc)
		}
	})

	t.Run("Error", func(t *testing.T) {
		b1, b2, b3 := new(bufWriter), &errWriter{errors.New("err")}, new(bufWriter)
		mw := xmlstream.MultiWriter(b1, b2, b3)

		tok := xml.CharData("test")
		if err := mw.EncodeToken(tok); err.Error() != "err" {
			t.Error("Write failed with unexpected error:", err)
		}

		if stok := string(tok); string(b1.b.(xml.CharData)) != stok || b3.b != nil {
			t.Errorf("One of the tokens is not correct. want=(%v,nil) got=(%v,%v)", tok, b1.b, b3.b)
		}
		if b1.wc != 1 || b3.wc != 0 {
			t.Errorf("Expected (1,0) writes, got: (%d,%d)", b1.wc, b3.wc)
		}
	})
}
