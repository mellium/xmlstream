// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.
//
// Code in this file copied from the Go io package:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE.GO file.

package xmlstream_test

import (
	"encoding/xml"
	"fmt"
	"io"
	"testing"
	"time"

	"mellium.im/xmlstream"
)

// TODO: This test package was a copy paste job with a mix of hand tweaking and
//       sed to get tests to pass. It should probably be rewritten from scratch
//       at some point.

func checkWrite(t *testing.T, w xmlstream.TokenWriter, tok xml.Token) {
	err := w.EncodeToken(tok)
	if err != nil {
		t.Errorf("write: %v", err)
	}
}

// Test a single read/write pair.
func TestPipe1(t *testing.T) {
	r, w := xmlstream.Pipe()
	go checkWrite(t, w, xml.CharData("hello, world"))
	tok, err := r.Token()
	if err != nil {
		t.Errorf("read: %v", err)
	} else if string(tok.(xml.CharData)) != "hello, world" {
		t.Errorf("bad read: got %s", tok)
	}
	r.Close()
	w.Close()
}

func reader(t *testing.T, r xml.TokenReader, c chan xml.Token) {
	for {
		tok, err := r.Token()
		if err == io.EOF {
			c <- nil
			break
		}
		if err != nil {
			t.Errorf("read: %v", err)
		}
		c <- tok
	}
}

// Test a sequence of read/write pairs.
func TestPipe2(t *testing.T) {
	c := make(chan xml.Token)
	r, w := xmlstream.Pipe()
	go reader(t, r, c)
	tok := xml.CharData("hello world")
	for i := 0; i < 5; i++ {
		err := w.EncodeToken(tok)
		if err != nil {
			t.Errorf("write: %v", err)
		}
		tt := <-c
		if string(tt.(xml.CharData)) != string(tok) {
			t.Errorf("wrote %v, read got %v", t, tt)
		}
	}
	w.Close()
	tt := <-c
	if tt != nil {
		t.Errorf("final read got %v", tt)
	}
}

// TODO: Test a large write that requires multiple reads to satisfy.

// Test read after/before writer close.

type closer interface {
	CloseWithError(error)
	Close() error
}

type pipeTest struct {
	async          bool
	err            error
	closeWithError bool
}

func (p pipeTest) String() string {
	return fmt.Sprintf("async=%v err=%v closeWithError=%v", p.async, p.err, p.closeWithError)
}

var pipeTests = []pipeTest{
	{true, nil, false},
	{true, nil, true},
	{true, io.ErrShortWrite, true},
	{false, nil, false},
	{false, nil, true},
	{false, io.ErrShortWrite, true},
}

func delayClose(t *testing.T, cl closer, ch chan struct{}, tt pipeTest) {
	time.Sleep(1 * time.Millisecond)
	var err error
	if tt.closeWithError {
		cl.CloseWithError(tt.err)
	} else {
		err = cl.Close()
	}
	if err != nil {
		t.Errorf("delayClose: %v", err)
	}
	ch <- struct{}{}
}

func TestPipeReadClose(t *testing.T) {
	for _, tt := range pipeTests {
		c := make(chan struct{}, 1)
		r, w := xmlstream.Pipe()
		if tt.async {
			go delayClose(t, w, c, tt)
		} else {
			delayClose(t, w, c, tt)
		}
		tok, err := r.Token()
		<-c
		want := tt.err
		if want == nil {
			want = io.EOF
		}
		if err != want {
			t.Errorf("read from closed pipe: %v want %v", err, want)
		}
		if tok != nil {
			t.Errorf("read on closed pipe returned %v", tok)
		}
		if err = r.Close(); err != nil {
			t.Errorf("r.Close: %v", err)
		}
	}
}

// Test close on Read side during Read.
func TestPipeReadClose2(t *testing.T) {
	c := make(chan struct{}, 1)
	r, _ := xmlstream.Pipe()
	go delayClose(t, r, c, pipeTest{})
	tok, err := r.Token()
	<-c
	if tok != nil || err != xmlstream.ErrClosedPipe {
		t.Errorf("read from closed pipe: %v, %v want %v, %v", tok, err, (xml.Token)(nil), xmlstream.ErrClosedPipe)
	}
}

// Test write after/before reader close.

func TestPipeWriteClose(t *testing.T) {
	for _, tt := range pipeTests {
		c := make(chan struct{}, 1)
		r, w := xmlstream.Pipe()
		if tt.async {
			go delayClose(t, r, c, tt)
		} else {
			delayClose(t, r, c, tt)
		}
		err := w.EncodeToken(xml.CharData("hello, world"))
		<-c
		expect := tt.err
		if expect == nil {
			expect = xmlstream.ErrClosedPipe
		}
		if err != expect {
			t.Errorf("write on closed pipe: %v want %v", err, expect)
		}
		if err = w.Close(); err != nil {
			t.Errorf("w.Close: %v", err)
		}
	}
}

// Test close on Write side during Write.
func TestPipeWriteClose2(t *testing.T) {
	c := make(chan struct{}, 1)
	_, w := xmlstream.Pipe()
	go delayClose(t, w, c, pipeTest{})
	err := w.EncodeToken(xml.CharData("hello, world"))
	<-c
	if err != xmlstream.ErrClosedPipe {
		t.Errorf("write to closed pipe: %v want %v", err, xmlstream.ErrClosedPipe)
	}
}

func TestWriteEmpty(t *testing.T) {
	r, w := xmlstream.Pipe()
	go func() {
		w.EncodeToken(xml.CharData("hello, world"))
		w.Close()
	}()

	r.Token()
	r.Close()
}

func TestWriteNil(t *testing.T) {
	r, w := xmlstream.Pipe()
	go func() {
		w.EncodeToken(nil)
		w.Close()
	}()
	r.Token()
	r.Close()
}

func TestWriteAfterWriterClose(t *testing.T) {
	r, w := xmlstream.Pipe()

	done := make(chan struct{})
	var writeErr error
	go func() {
		err := w.EncodeToken(xml.CharData("hello"))
		if err != nil {
			t.Errorf("got error: %q; expected none", err)
		}
		w.Close()
		writeErr = w.EncodeToken(xml.CharData("world"))
		done <- struct{}{}
	}()

	var result string
	tok, err := r.Token()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	result = string(tok.(xml.CharData))
	<-done

	if result != "hello" {
		t.Errorf("got: %q; want: %q", result, "hello")
	}
	if writeErr != xmlstream.ErrClosedPipe {
		t.Errorf("got: %q; want: %q", writeErr, xmlstream.ErrClosedPipe)
	}
}
