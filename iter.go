// Copyright 2019 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Iter provides a mechanism for iterating over the children of an XML element.
// Successive calls to the Next method will step through each child, returning
// its start element and a reader that is limited to the remainder of the child.
type Iter struct {
	r       TokenReadCloser
	err     error
	next    *xml.StartElement
	cur     xml.TokenReader
	closed  bool
	discard TokenWriter
}

type nopClose struct{}

func (nopClose) Close() error { return nil }

func wrapClose(r xml.TokenReader) TokenReadCloser {
	var c io.Closer
	var ok bool
	c, ok = r.(io.Closer)
	if !ok {
		c = nopClose{}
	}

	return struct {
		xml.TokenReader
		io.Closer
	}{
		TokenReader: Inner(r),
		Closer:      c,
	}
}

// NewIter returns a new iterator that iterates over the children of the most
// recent start element already consumed from r.
func NewIter(r xml.TokenReader) *Iter {
	iter := &Iter{
		r:       wrapClose(r),
		discard: Discard(),
	}
	return iter
}

// Next returns true if there are more items to decode.
func (i *Iter) Next() bool {
	if i.err != nil || i.closed {
		return false
	}

	// Consume the previous element before moving on to the next.
	if i.cur != nil {
		_, i.err = Copy(i.discard, i.cur)
		if i.err != nil {
			return false
		}
	}

	i.next = nil
	t, err := i.r.Token()
	if err != nil {
		if err != io.EOF {
			i.err = err
		}
		return false
	}

	if start, ok := t.(xml.StartElement); ok {
		i.next = &start
		i.cur = MultiReader(Inner(i.r), Token(i.next.End()))
		return true
	}
	return false
}

// Current returns a reader over the most recent child.
func (i *Iter) Current() (*xml.StartElement, xml.TokenReader) {
	return i.next, i.cur
}

// Err returns the last error encountered by the iterator (if any).
func (i *Iter) Err() error {
	return i.err
}

// Close indicates that we are finished with the given iterator.
// Calling it multiple times has no effect.
//
// If the underlying TokenReader is also an io.Closer, Close calls the readers
// Close method.
func (i *Iter) Close() error {
	if i.closed {
		return nil
	}

	i.closed = true
	_, err := Copy(i.discard, i.r)
	if err != nil {
		return err
	}
	return i.r.Close()
}
