// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"errors"
	"io"
)

// Wrap wraps a token stream in a start element and its corresponding end
// element.
func Wrap(r TokenReader, start xml.StartElement) TokenReader {
	state := 0
	return ReaderFunc(func() (t xml.Token, err error) {
		switch state {
		case 0:
			state++
			return start, nil
		case 1:
			t, err = r.Token()
			if err == io.EOF {
				state++
				err = nil
			}
			if t == nil {
				state++
				t = start.End()
			}
			return t, err
		case 2:
			state++
			return start.End(), io.EOF
		default:
			return nil, io.EOF
		}
	})
}

// Unwrap returns a new token reader that skips the first token read from it
// and, if it is a start element, also skips its corresponding end element.
// If the element is not a start element it is returned along with an error.
func Unwrap(r TokenReader) TokenReader {
	var gotStart bool
	var ok bool
	var start xml.StartElement
	depth := 0
	return ReaderFunc(func() (t xml.Token, err error) {
		if !gotStart {
			t, err = r.Token()
			gotStart = true
			start, ok = t.(xml.StartElement)
			if err == io.EOF {
				err = nil
			}
			if ok {
				t = nil
			}
			if !ok && err == nil {
				err = errors.New("xmlstream: unwrap expected start element")
				return t, err
			}
		}

		t, err = r.Token()
		switch tok := t.(type) {
		case xml.StartElement:
			if tok.Name == start.Name {
				depth++
			}
		case xml.EndElement:
			if tok.Name == start.Name {
				depth--
				if depth == -1 {
					t, err = r.Token()
				}
			}
		}

		return t, err
	})
}
