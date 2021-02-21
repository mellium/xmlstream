// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.
//
// Code in this file copied from the Go io package:
//
// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.GO file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

type eofReader struct{}

func (eofReader) Token() (xml.Token, error) {
	return nil, io.EOF
}

type multiReader struct {
	readers []xml.TokenReader
}

func (mr *multiReader) Token() (t xml.Token, err error) {
	for len(mr.readers) > 0 {
		// Optimization to flatten nested multiReaders (https://golang.org/issue/13558).
		if len(mr.readers) == 1 {
			if r, ok := mr.readers[0].(*multiReader); ok {
				mr.readers = r.readers
				continue
			}
		}
		if mr.readers[0] == nil {
			mr.readers = mr.readers[1:]
			continue
		}
		t, err = mr.readers[0].Token()
		if err == io.EOF || (t == nil && err == nil) {
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (https://golang.org/issue/18232).
			mr.readers[0] = eofReader{} // permit earlier GC
			mr.readers = mr.readers[1:]
		}
		if t != nil || (err != nil && err != io.EOF) {
			if err == io.EOF && len(mr.readers) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return nil, io.EOF
}

// MultiReader returns an xml.TokenReader that's the logical concatenation of
// the provided input readers.
// They're read sequentially.
// Once all inputs have returned io.EOF, Token will return io.EOF.
// If any of the readers return a non-nil, non-EOF error, Token will return that
// error.
func MultiReader(readers ...xml.TokenReader) xml.TokenReader {
	r := make([]xml.TokenReader, len(readers))
	copy(r, readers)
	return &multiReader{readers: readers}
}

type multiWriter struct {
	writers []TokenWriter
}

func (mw *multiWriter) EncodeToken(t xml.Token) error {
	for _, w := range mw.writers {
		if err := w.EncodeToken(t); err != nil {
			return err
		}
	}
	return nil
}

func (mw *multiWriter) Flush() error {
	for _, w := range mw.writers {
		flusher, ok := w.(Flusher)
		if !ok {
			continue
		}
		if err := flusher.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
// If any of the writers return an error, the MultiWriter immediately returns
// the error and stops writing.
func MultiWriter(writers ...TokenWriter) TokenWriter {
	w := make([]TokenWriter, len(writers))
	copy(w, writers)
	return &multiWriter{w}
}
