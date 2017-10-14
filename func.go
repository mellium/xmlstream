// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
)

// ReaderFunc type is an adapter to allow the use of ordinary functions as a
// TokenReader. If f is a function with the appropriate signature,
// ReaderFunc(f) is an TokenReader that calls f.
type ReaderFunc func() (xml.Token, error)

// Token calls f.
func (f ReaderFunc) Token() (xml.Token, error) {
	return f()
}

// WriterFunc type is an adapter to allow the use of ordinary functions as a
// TokenWriter with a nop Flush method. If f is a function with the appropriate
// signature, WriterFunc(f) is an TokenWriter that calls f.
type WriterFunc func(t xml.Token) error

// EncodeToken calls f.
func (f WriterFunc) EncodeToken(t xml.Token) error {
	return f(t)
}

// Flush is a nop.
func (WriterFunc) Flush() error {
	return nil
}
